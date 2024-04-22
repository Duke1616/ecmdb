package mongox

type Collection interface {
	Where()
}

type Coll struct {
	collName string // 集合名
	*Mongo
}

func (m *Coll) Where() {

}
