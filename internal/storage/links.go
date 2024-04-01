package storage

type link struct {
	shortLink string
	fullLink  string
}

type Links []link

func (links *Links) Add(sl, fl string) {
	*links = append(*links, link{shortLink: sl, fullLink: fl})
}

func (links *Links) Get(shortLink string) (string, bool) {
	for _, link := range *links {
		if link.shortLink == shortLink {
			return link.fullLink, true
		}
	}
	return "", false
}
