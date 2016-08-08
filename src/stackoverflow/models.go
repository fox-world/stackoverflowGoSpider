package stackoverflow

import(
"time"
)

type Post struct {
	Title        string
	Link         string
	Postuser     string
	Postuserlink string
	Posttime     time.Time
	Vote         int
	Viewed       int
}

