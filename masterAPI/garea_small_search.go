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
)

type GAreaSmallSearchResp struct {
	Garea_small []Area_s `json:"garea_small"`
}
type Area_s struct {
	Code    string `json:"areacode_s"`
	Name    string `json:"areaname_s"`
	Garea_m Area_m `json:garea_middle`
	Garea_l Area_l `json:garea_large`
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
	for _, ga_small := range gresp.Garea_small {
		if ifNotFound(db, `select count(*) as count from area_m where code = $1`, ga_small.Garea_m.Code) {
			_, err := db.Exec("insert into area_m values($1, $2)",
				ga_small.Garea_m.Code, ga_small.Garea_m.Name)
			if err != nil {
				fmt.Println(1)
				log.Fatal(err)
			}
		}
		if ifNotFound(db, `select count(*) as count from area_l where code = $1`, ga_small.Garea_l.Code) {
			_, err = db.Exec("insert into area_l values($1, $2)",
				ga_small.Garea_l.Code, ga_small.Garea_l.Name)
			if err != nil {
				fmt.Println(2)
				log.Fatal(err)
			}
		}
		if ifNotFound(db, `select count(*) as count from pref where code = $1`, ga_small.Pref.Code) {
			_, err = db.Exec("insert into pref values($1, $2)",
				ga_small.Pref.Code, ga_small.Pref.Name)
			if err != nil {
				log.Fatal(err)
			}
		}
		_, err = db.Exec("insert into area_s values($1, $2, $3, $4, $5)",
			ga_small.Code, ga_small.Name, ga_small.Garea_m.Code, ga_small.Garea_l.Code, ga_small.Pref.Code)
		if err != nil {
			log.Fatal(err)
		}
	}
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
