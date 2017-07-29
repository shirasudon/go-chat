package entity

import (
	"errors"
)

// Repositories holds any XXXRepository.
// user can get each repository from this.
type Repositories interface {
	Users() UserRepository
	Messages() MessageRepository

	// finalize database connection
	Close() error
}

// producer function for generating Repositories.
// it is used to generate specific Repositories in OpenRepositories.
// the driverSourceName is typically dabase file name such as database.db,
// or database.sqlite in sqlite.
var RepositoryProducer func(driverSourceName string) (Repositories, error)

// Repo holds any Repositories.
// it is initialized by importing other packages.
// example:
//  import (
//    github.com/mzki/chat/entity
//    _ github.com/mzki/chat/entity/stub
//  )
//
//  // open resistered Repositories.
//  repos, err := entity.OpenRepositories("database.db")
//
//  // we can get stub repository from entity.Repos.
//  // since it is initialized by importing github.com/mzki/chat/entity/stub
//  userRepos := repos.Users()
//
//  repositories is cached and return it when OpenRepositories is called twice or above.
func OpenRepositories(driverSourceName string) (Repositories, error) {
	if repositories != nil {
		return repositories, nil
	}

	// check exsitance of producer function and execute it.
	if RepositoryProducer == nil {
		return nil, errors.New(`entity.OpenRepositories: no exists for Repositories Producer.
you should import producer package such as github/mzki/chat/entity/stub`)
	}
	var err error
	repositories, err = RepositoryProducer(driverSourceName)
	return repositories, err
}

// cache for the Repositories
var repositories Repositories

func mustRepositories() {
	if repositories == nil {
		panic("No initialized Repositories. You should call OpenRepositories firstly")
	}
}

// return UserRepository from initialized Repositories.
// Be sure to call OpenRepositories() before use this.
func Users() UserRepository {
	mustRepositories()
	return repositories.Users()
}

// return MessagesRepository from initialized Repositories.
// Be sure to call OpenRepositories() before use this.
func Messages() MessageRepository {
	mustRepositories()
	return repositories.Messages()
}
