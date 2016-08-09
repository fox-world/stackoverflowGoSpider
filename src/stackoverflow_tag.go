package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"labix.org/v2/mgo"
	"log"
	"os"
	"stackoverflow"
	"strconv"
	"strings"
	"time"
)

const CONCURRENT_SIZE = 10

var session mgo.Session

func main() {
	logFileName := "test_" + time.Now().Format("2006_01_02") + ".log"
	logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("error opening file: %v", err)
	}
	defer logFile.Close()

	//log.SetOutput(logFile)

	starttime := time.Now()
	parseTag("go")
	endtime := time.Now()
	timecost := endtime.Unix() - starttime.Unix()

	log.Println("=====Begin time:\t", starttime.Format("2006-01-02 15:04:05"))
	log.Println("=====End time:\t", endtime.Format("2006-01-02 15:04:05"))
	log.Println("=====Total time cost:\t", timecost)
}

func parseTag(tag string) {

	dburi := "mongodb://admin:123456@localhost/stackoverflow"
	session, err := mgo.Dial(dburi)
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	if err != nil {
		panic(err)
	}
	postsCollection := session.DB("stackoverflow").C("posts")

	url := "http://stackoverflow.com/questions/tagged/" + tag
	totalPage := queryTotalPage(url + "?page=1&sort=newest&pagesize=50")

	chs := make([]chan string, CONCURRENT_SIZE)
	for i := 0; i < CONCURRENT_SIZE; i++ {
		chs[i] = make(chan string, 1)
	}

	var pageurl string
	for i := 1; i <= totalPage; i++ {
		pageurl = url + "?page=" + strconv.Itoa(i) + "&sort=newest&pagesize=50"
		go parseQuestions(pageurl, chs[(i-1)%CONCURRENT_SIZE], postsCollection)
		if i%CONCURRENT_SIZE == 0 {
			clearChannel(chs, CONCURRENT_SIZE)
		}
	}

	var leftCount = totalPage % CONCURRENT_SIZE
	clearChannel(chs, leftCount)
}

func clearChannel(chs []chan string, size int) {
	var url string
	for i := 0; i < size; i++ {
		url = <-chs[i]
		log.Println("++++++++++Finished parsing page:\t", url)
	}
}

func parseQuestions(url string, ch chan string, pCollection *mgo.Collection) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	ch <- url
	log.Println("++++++++++Parsing page:\t", url)

	var posts []interface{}

	doc.Find(".question-summary").Each(func(i int, s *goquery.Selection) {

		posttimestr, _ := s.Find(".relativetime").Attr("title")

		linkSelection := s.Find(".summary>h3>a")
		title := strings.TrimSpace(linkSelection.Text())
		link, _ := linkSelection.Attr("href")

		votestr := s.Find(".vote-count-post>strong").Text()
		viewstr := strings.TrimSpace(s.Find(".views").Text())
		viewstr = strings.Split(viewstr, " ")[0]

		userdetails := s.Find(".user-details>a")
		username := strings.TrimSpace(userdetails.Text())
		userlink, _ := userdetails.Attr("href")

		layout := "2014-09-12 11:45:26.37Z"
		posttime, _ := time.Parse(layout, posttimestr)
		vote, _ := strconv.Atoi(votestr)
		views, _ := strconv.Atoi(viewstr)

		post := stackoverflow.Post{Title: title, Link: link, Postuser: username, Postuserlink: userlink, Posttime: posttime, Vote: vote, Viewed: views}
		posts = append(posts, post)

		log.Println("-------------------------------------------------------------------")
		log.Println("post time:", posttime)
		log.Println("user name:", username)
		log.Println("user link:", userlink)
		log.Println("vote:", vote)
		log.Println("views:", views)
		log.Println("title:", title)
		log.Println("link:", link)
	})

    err = pCollection.Insert(posts...)
	if err != nil {
		panic(err)
	} else {
		log.Println("-------------insert ",len(posts)," psots success-------------------")
	}
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
