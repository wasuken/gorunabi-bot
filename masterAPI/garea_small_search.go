package masterAPI

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"unicode/utf8"
)

// db
type Area_s_db struct {
	Code         string
	Name         string
	Garea_m_code string
	Garea_l_code string
	Pref_code    string
}

// api
type GAreaSmallSearchResp struct {
	Garea_small []Area_s `json:"garea_small"`
}
type Area_s struct {
	Code    string `json:"areacode_s"`
	Name    string `json:"areaname_s"`
	Garea_m Area_m `json:"garea_middle"`
	Garea_l Area_l `json:"garea_large"`
	Pref    Pref   `json:"pref"`
}
type Area_m struct {
	Code string `json:"areacode_m"`
	Name string `json:"areaname_m"`
}
type Area_l struct {
	Code string `json:"areacode_l"`
	Name string `json:"areaname_l"`
}
type Pref struct {
	Code string `json:"pref_code"`
	Name string `json:"pref_name"`
}

var (
	db_name_list        [4]string         = [...]string{"area_s", "area_m", "area_l", "pref"}
	db_name_api_key_map map[string]string = map[string]string{"area_s": "areacode_s",
		"area_m": "areacode_m", "area_l": "areacode_l", "pref": "pref"}
)

// keywordから、地名を算出し、適切なareaのパラメータを生成する
func SearchMasterDataMakeKeyValues(keyword string) map[string]string {
	var kvs map[string]string
	keyword_split := []string{}
	for _, k := range strings.Split(keyword, " ") {
		if utf8.RuneCountInString(k) >= 2 {
			keyword_split = append(keyword_split, k)
		}
	}
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	name_and_code_list_map := allMasterNameMap(db, keyword_split)
	for key, name_and_code_list := range name_and_code_list_map {
		for _, result := range name_and_code_list {
			if len(result) > 0 {
				// 更新されていくのでとりあえず重複の心配はない
				kvs[db_name_api_key_map[key]] = result[1]
			}
		}
	}
	return kvs
}
func allMasterNameMap(db *sql.DB, keywords []string) map[string][][]string {
	whereStr := ""
	for _, keyword := range keywords {
		if whereStr != "" {
			whereStr += " or "
		}

		whereStr += "name like " + "'%" + template.HTMLEscapeString(keyword) + "%'"
	}
	whereStr = "where " + whereStr

	name_list_map := make(map[string][][]string)
	for _, v := range db_name_list {
		base_sql := fmt.Sprintf("select name, code from %s ", v) + whereStr
		rows, e := db.Query(base_sql)
		if e != nil {
			log.Fatal(e)
		}
		var name, code string
		for rows.Next() {
			er := rows.Scan(&name, &code)
			if er != nil {
				log.Fatal(er)
			}
			name_list_map[v] = append(name_list_map[v], []string{name, code})
		}
		defer rows.Close()
	}
	return name_list_map
}

// マスタの取得
func GetGAreaSmallSearchResponse(api_base_url string) {
	resp, _ := http.Get(fmt.Sprintf("%s/%s/%s/?keyid=%s&lang=ja", api_base_url,
		"master", "GAreaSmallSearchAPI/v3", os.Getenv("GURUNABI_SECRET")))
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	var gresp GAreaSmallSearchResp
	if err := json.Unmarshal(byteArray, &gresp); err != nil {
		log.Fatal(err)
	}
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("delete from area_l")
	db.Exec("delete from area_m")
	db.Exec("delete from area_s")
	db.Exec("delete from pref")
	// DBにデータを入れる。
	bulkInsertStr := ""
	var area_m_codes, area_l_codes, pref_codes []string
	for _, ga_small := range gresp.Garea_small {
		if !contains(area_m_codes, ga_small.Garea_m.Code) {
			bulkInsertStr += fmt.Sprintf(`insert into area_m values('%s', '%s');`,
				ga_small.Garea_m.Code, ga_small.Garea_m.Name)
			area_m_codes = append(area_m_codes, ga_small.Garea_m.Code)
		}
		if !contains(area_l_codes, ga_small.Garea_l.Code) {
			bulkInsertStr += fmt.Sprintf(`insert into area_l values('%s', '%s');`,
				ga_small.Garea_l.Code, ga_small.Garea_l.Name)
			area_l_codes = append(area_l_codes, ga_small.Garea_l.Code)
		}
		if !contains(pref_codes, ga_small.Pref.Code) {
			bulkInsertStr += fmt.Sprintf(`insert into pref values('%s', '%s');`,
				ga_small.Pref.Code, ga_small.Pref.Name)
			pref_codes = append(pref_codes, ga_small.Pref.Code)
		}
		bulkInsertStr += fmt.Sprintf(`insert into area_s values('%s', '%s', '%s', '%s', '%s');`,
			ga_small.Code, ga_small.Name, ga_small.Garea_m.Code, ga_small.Garea_l.Code, ga_small.Pref.Code)
	}
	fmt.Println("created str")
	_, err = db.Exec(bulkInsertStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("finish")
}

func contains(ary []string, value string) bool {
	for _, v := range ary {
		if v == value {
			return true
		}
	}
	return false
}

func ifNotFound(db *sql.DB, sqlStr, code string) bool {
	count := 0
	err := db.QueryRow(sqlStr, code).Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	return count > 0
}

// マスタの作成
func CreateTables() {
	// DBにデータを入れる。
	f, err := os.Open("create.sql")
	if err != nil {
		log.Fatal("error")
	}
	defer f.Close()
	// 一気に全部読み取り
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("error")
	}
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	defer db.Close()
	if err != nil {
		log.Fatal(err)
	}
	db.Query(string(b))
}
