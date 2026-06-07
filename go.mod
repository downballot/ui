module github.com/downballot/ui

go 1.26.2

require (
	github.com/downballot/downballot v0.0.0-00010101000000-000000000000
	github.com/go-app-blazar/router v0.1.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	github.com/lmittmann/tint v1.1.3
	github.com/mattn/go-isatty v0.0.22
	github.com/maxence-charriere/go-app/v11 v11.0.4
	github.com/tekkamanendless/httperror v1.0.1
	github.com/tekkamanendless/restapiclient v0.1.1
)

require (
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/tekkamanendless/restfulwrapper v0.2.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	gorm.io/gorm v1.31.1 // indirect
)

replace github.com/downballot/downballot => ../downballot

//replace github.com/maxence-charriere/go-app/v11 => github.com/tekkamanendless/fork-of-maxence-charriere-go-app/v11 v11.0.0-20260602191626-433345bdb528

//replace github.com/maxence-charriere/go-app/v11 => ../../tekkamanendless/fork-of-maxence-charriere-go-app
