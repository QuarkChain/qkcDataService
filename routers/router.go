package routers

import (
	"github.com/QuarkChain/qkcDataService/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
}
