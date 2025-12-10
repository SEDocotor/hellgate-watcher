module github.com/SEDocotor/hellgate-watcher-go

go 1.20

require (
	gorm.io/driver/postgres v1.4.6
	gorm.io/gorm v1.31.1
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.2.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/text v0.20.0 // indirect
)

// Workaround: map gorm.io module paths to the canonical GitHub modules at chosen versions
replace gorm.io/gorm => github.com/go-gorm/gorm v1.31.1

replace gorm.io/driver/postgres => github.com/go-gorm/postgres v1.4.6
