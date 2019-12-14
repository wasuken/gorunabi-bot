package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type RestAPIResp struct {
	Rest []Rest `json:"rest"`
}
type Rest struct {
	Name      string `json:"name"`
	UrlMobile string `json:"url_mobile"`
}

// レストラン検索の想定
func GetGurunabiJSONResult(api_base_url, paramsStr string) string {
	resp, _ := http.Get(api_base_url + "/RestSearchAPI/v3/?keyid=" +
		os.Getenv("GURUNABI_SECRET") + "&" + paramsStr)
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)

	result := ""
	var restJsonApiResp RestAPIResp
	if err := json.Unmarshal(byteArray, &restJsonApiResp); err != nil {
		log.Fatal(err)
	}
	for _, rest := range restJsonApiResp.Rest {
		result += rest.Name + "\n" +
			rest.UrlMobile + "\n"
	}
	return result
}
