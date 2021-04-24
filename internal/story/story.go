package story

var dirName = "./stories"

type Story struct {
	ID      int64
	Title   string
	Content []ContentPart
	Parts   int
}

type ContentPart struct {
	Image   string
	Caption string
}
