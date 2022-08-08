package media

type ExEpisodeType uint8

func (exEpisodeType ExEpisodeType) String() string {
	return ExEpisodeTypes[exEpisodeType]
}

const (
	OTHER ExEpisodeType = iota
	HALF
	SP
	OVA
	OAD
	MOVIE
	OFFSET
)
const (
	OTHER32 uint32 = iota
	HALF32
	SP32
	OVA32
	OAD32
	MOVIE32
	OFFSET32
)

var (
	ExEpisodeTypes []string = []string{"其他", "", "SP", "OVA", "OAD", "劇場版", ""}
)
