package RockInterface

type ICommand interface {
	Run([]string) string
	Help() string
	Name() string
}
