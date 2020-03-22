package txcoursecrawler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

//发送api请求
func apiRequest(url string, body io.Reader, ret interface{}) error {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, body)
	// 自定义Header
	//req.Header.Set("User-Agent", "Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.87 Safari/537.36")
	req.Header.Set("referer", "https://fudao.qq.com")
	var resp *http.Response
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("get response %s %q", resp.Status, respBody)
	}
	if err = json.Unmarshal(respBody, ret); err != nil {
		return err
	}
	return nil
}
