package storage

type link struct {
	idLink   string
	fullLink string
}

type Links []link

func (links *Links) Add(sl, fl string) {
	*links = append(*links, link{idLink: sl, fullLink: fl})
}

func (links *Links) Get(idLink string) (string, bool) {
	for _, link := range *links {
		if link.idLink == idLink {
			return link.fullLink, true
		}
	}
	return "", false
}
