package ports

type (
	RepositoriesFactory interface {
		LoginRepository() LoginRepository
		PokerRepository() PokerRepository
	}
)
