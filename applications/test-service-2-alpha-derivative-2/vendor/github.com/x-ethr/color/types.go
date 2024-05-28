package color

type escape interface {
	typecast(entity interface{}) string
}
