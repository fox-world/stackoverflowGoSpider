package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strconv"
	"strings"
)

func main() {
	url := "http://stackoverflow.com/questions/tagged/go"
	parseTag(url)
}

func parseTag(url string) {
	totalPage := queryTotalPage(url + "?page=1&sort=newest&pagesize=50")

	var pageurl string
	for i := 1; i <= totalPage; i++ {
		pageurl = url + "?page=" + strconv.Itoa(i) + "&sort=newest&pagesize=50"
		log.Println("++++++++++Parsing page:\t", pageurl)
		parseQuestions(pageurl)
	}
}

func parseQuestions(url string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

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

		fmt.Println("-------------------------------------------------------------------")
		fmt.Println("post time:", posttime)
		fmt.Println("user name:", username)
		fmt.Println("user link:", userlink)
		fmt.Println("vote:", vote)
		fmt.Println("views:", views)
		fmt.Println("title:", title)
		fmt.Println("link:", link)
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
