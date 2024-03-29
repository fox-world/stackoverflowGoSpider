package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"os"
	"stack-spider/stackoverflow"
	"strconv"
	"strings"
	"time"
)

const CONCURRENT_SIZE = 10

var session mgo.Session
var totalRecord int
var status stackoverflow.Status

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
	log.Println("=====Total record:\t", totalRecord)
}

func parseTag(tag string) {

	status.UpdateStatus(true)

	dbUri := "mongodb://admin:123456@localhost/stackoverflow"
	session, err := mgo.Dial(dbUri)
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	if err != nil {
		panic(err)
	}
	postsCollection := session.DB("stackoverflow").C("posts")

	url := "http://stackoverflow.com/questions/tagged/" + tag
	totalPage := queryTotalPage(url + "?page=1&sort=newest&pagesize=50")

	chs := make([]chan int, CONCURRENT_SIZE)
	for i := 0; i < CONCURRENT_SIZE; i++ {
		chs[i] = make(chan int, 1)
	}

	var pageUrl string
	for i := 1; i <= totalPage; i++ {
		if status.IsRun() {
			pageUrl = url + "?page=" + strconv.Itoa(i) + "&sort=newest&pagesize=50"
			go parseQuestions(pageUrl, chs[(i-1)%CONCURRENT_SIZE], postsCollection)
			if i%CONCURRENT_SIZE == 0 {
				clearChannel(chs, CONCURRENT_SIZE)
			}
		} else {
			break
		}

	}

	if status.IsRun() {
		var leftCount = totalPage % CONCURRENT_SIZE
		clearChannel(chs, leftCount)
	}

}

func clearChannel(chs []chan int, size int) {
	for i := 0; i < size; i++ {
		totalRecord += <-chs[i]
	}
}

func parseQuestions(url string, ch chan int, pCollection *mgo.Collection) {

	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("++++++++++Parsing page:\t", url)

	var posts []interface{}

	doc.Find(".question-summary").Each(func(i int, s *goquery.Selection) {

		postTimeStr, _ := s.Find(".relativetime").Attr("title")

		linkSelection := s.Find(".summary>h3>a")
		title := strings.TrimSpace(linkSelection.Text())
		link, _ := linkSelection.Attr("href")

		voteStr := s.Find(".vote-count-post>strong").Text()
		viewStr := strings.TrimSpace(s.Find(".views").Text())
		viewStr = strings.Split(viewStr, " ")[0]

		var tags []string
		s.Find(".summary>.tags>.post-tag").Each(func(j int, t *goquery.Selection) {
			tags = append(tags, strings.TrimSpace(t.Text()))
		})

		userDetails := s.Find(".user-details>a")
		userName := strings.TrimSpace(userDetails.Text())
		userLink, _ := userDetails.Attr("href")

		layout := "2006-01-02 15:04:05Z"
		postTime, _ := time.Parse(layout, postTimeStr)
		vote, _ := strconv.Atoi(voteStr)
		views, _ := strconv.Atoi(viewStr)

		dbPost := stackoverflow.Post{}
		err = pCollection.Find(bson.M{"title": title, "postTime": postTime, "link": link}).One(&dbPost)
		if err != nil {
			log.Println("+++++++++++++++add new post:\t", title)
			post := stackoverflow.Post{Title: title, Link: link, Tags: tags, PostUser: userName, PostUserLink: userLink, PostTime: postTime, Vote: vote, Viewed: views}
			posts = append(posts, post)

			log.Println("-------------------------------------------------------------------")
			log.Println("post time:", postTime)
			log.Println("user name:", userName)
			log.Println("user link:", userLink)
			log.Println("tags:", tags)
			log.Println("vote:", vote)
			log.Println("views:", views)
			log.Println("title:", title)
			log.Println("link:", link)
		} else {
			log.Println("------------exists post:\t", title)
			if status.IsRun() {
				status.UpdateStatus(false)
			}
		}

	})

	if len(posts) > 0 {
		err = pCollection.Insert(posts...)
		if err != nil {
			panic(err)
		} else {
			log.Println("-------------insert ", len(posts), " psots success for url:", url)
		}
	}

	ch <- len(posts)
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
