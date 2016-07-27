package main

import (
	"github.com/PuerkitoBio/goquery"
	"log"
	"strconv"
	"strings"
	"time"
)

func main() {
	url := "http://stackoverflow.com/questions/tagged/go"
	starttime := time.Now()
	parseTag(url)
	endtime := time.Now()
	timecost := endtime.Unix() - starttime.Unix()

	log.Println("=====Begin time:\t", starttime.Format("2006-01-02 15:04:05"))
	log.Println("=====End time:\t", endtime.Format("2006-01-02 15:04:05"))
	log.Println("=====Total time cost:\t", timecost)
}

func parseTag(url string) {
	totalPage := queryTotalPage(url + "?page=1&sort=newest&pagesize=50")

	chs := make([]chan string, 10)
	for i := 0; i < 10; i++ {
		chs[i] = make(chan string, 1)
	}

	var pageurl string
	for i := 1; i <= totalPage; i++ {
		pageurl = url + "?page=" + strconv.Itoa(i) + "&sort=newest&pagesize=50"
		go parseQuestions(pageurl, chs[(i-1)%10])
		if i%10 == 0 {
			clearChannel(chs, 10)
		}
	}

	var leftCount = totalPage % 10
	clearChannel(chs, leftCount)
}

func clearChannel(chs []chan string, size int) {
	var url string
	for i := 0; i < size; i++ {
		url = <-chs[i]
		log.Println("++++++++++Finished parsing page:\t", url)
	}
}

func parseQuestions(url string, ch chan string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	ch <- url
	log.Println("++++++++++Parsing page:\t", url)

	doc.Find(".question-summary").Each(func(i int, s *goquery.Selection) {

		posttime, _ := s.Find(".relativetime").Attr("title")

		linkSelection := s.Find(".summary>h3>a")
		title := strings.TrimSpace(linkSelection.Text())
		link, _ := linkSelection.Attr("href")

		vote := s.Find(".vote-count-post>strong").Text()
		views := strings.TrimSpace(s.Find(".views").Text())
		views = strings.Split(views, " ")[0]

		userdetails := s.Find(".user-details>a")
		username := strings.TrimSpace(userdetails.Text())
		userlink, _ := userdetails.Attr("href")

		log.Println("-------------------------------------------------------------------")
		log.Println("post time:", posttime)
		log.Println("user name:", username)
		log.Println("user link:", userlink)
		log.Println("vote:", vote)
		log.Println("views:", views)
		log.Println("title:", title)
		log.Println("link:", link)
	})
}

func queryTotalPage(url string) int {
	totalPage := 0
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	next := doc.Find(".page-numbers.next")
	prev := next.Parent().Prev()

	totalPageStr := strings.TrimSpace(prev.Find("span.page-numbers").Text())
	totalPage, _ = strconv.Atoi(totalPageStr)

	return totalPage
}
