package routes

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/apex/log"
	"github.com/crawlab-team/crawlab-core/controllers"
	"github.com/crawlab-team/crawlab-core/web"
	"github.com/gin-gonic/gin"
)

type RouterServiceInterface interface {
	RegisterControllerToGroup(group *gin.RouterGroup, basePath string, ctr controllers.ListController)
	RegisterHandlerToGroup(group *gin.RouterGroup, path string, method string, handler gin.HandlerFunc)
}

type RouterService struct {
	app *gin.Engine
}

func NewRouterService(app *gin.Engine) (svc *RouterService) {
	return &RouterService{
		app: app,
	}
}

func (svc *RouterService) RegisterControllerToGroup(group *gin.RouterGroup, basePath string, ctr controllers.BasicController) {
	group.GET(basePath, ctr.Get)
	group.POST(basePath, ctr.Post)
	group.PUT(basePath, ctr.Put)
	group.DELETE(basePath, ctr.Delete)
}

func (svc *RouterService) RegisterListControllerToGroup(group *gin.RouterGroup, basePath string, ctr controllers.ListController) {
	group.GET(basePath+"/:id", ctr.Get)
	group.GET(basePath, ctr.GetList)
	group.POST(basePath, ctr.Post)
	group.POST(basePath+"/batch", ctr.PostList)
	group.PUT(basePath+"/:id", ctr.Put)
	group.PUT(basePath, ctr.PutList)
	group.DELETE(basePath+"/:id", ctr.Delete)
	group.DELETE(basePath, ctr.DeleteList)
}

func (_ *RouterService) RegisterActionControllerToGroup(group *gin.RouterGroup, basePath string, ctr controllers.ActionController) {
	for _, action := range ctr.Actions() {
		routerPath := path.Join(basePath, action.Path)
		switch action.Method {
		case http.MethodGet:
			group.GET(routerPath, action.HandlerFunc)
		case http.MethodPost:
			group.POST(routerPath, action.HandlerFunc)
		case http.MethodPut:
			group.PUT(routerPath, action.HandlerFunc)
		case http.MethodDelete:
			group.DELETE(routerPath, action.HandlerFunc)
		}
	}
}

func (svc *RouterService) RegisterListActionControllerToGroup(group *gin.RouterGroup, basePath string, ctr controllers.ListActionController) {
	svc.RegisterListControllerToGroup(group, basePath, ctr)
	svc.RegisterActionControllerToGroup(group, basePath, ctr)
}

func (svc *RouterService) RegisterHandlerToGroup(group *gin.RouterGroup, path string, method string, handler gin.HandlerFunc) {
	switch method {
	case http.MethodGet:
		group.GET(path, handler)
	case http.MethodPost:
		group.POST(path, handler)
	case http.MethodPut:
		group.PUT(path, handler)
	case http.MethodDelete:
		group.DELETE(path, handler)
	default:
		log.Warn(fmt.Sprintf("%s is not a valid http method", method))
	}
}

func InitRoutes(app *gin.Engine) (err error) {
	// routes groups
	groups := NewRouterGroups(app)

	// router service
	svc := NewRouterService(app)

	// register routes
	registerRoutesAnonymousGroup(svc, groups)
	registerRoutesAuthGroup(svc, groups)
	registerRoutesFilterGroup(svc, groups)
	registerStaticRoutes(svc, groups)

	return nil
}

func registerRoutesAnonymousGroup(svc *RouterService, groups *RouterGroups) {
	// login
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/", controllers.LoginController)

	// version
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/version", controllers.VersionController)

	// i18n
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/i18n", controllers.I18nController)

	// system info
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/system-info", controllers.SystemInfoController)

	// demo
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/demo", controllers.DemoController)
}

func registerStaticRoutes(svc *RouterService, groups *RouterGroups) error {
	distFS, err := web.DistFS()
	if err != nil {
		return err
	}

	svc.app.NoRoute(readHandler(distFS))
	return nil
}

func readHandler(distFS fs.FS) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p := strings.TrimPrefix(ctx.Request.URL.Path, "/")
		p = path.Clean(p)

		if p != "." {
			if data, ok := readFile(distFS, p); ok {
				ctx.Data(http.StatusOK, contentType(p), data)
				return
			}
		}

		index, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.Data(http.StatusOK, "text/html; charset=utf-8", index)
	}
}

func readFile(distFS fs.FS, name string) ([]byte, bool) {
	_, err := fs.Stat(distFS, name)
	if err != nil {
		return nil, false
	}

	data, err := fs.ReadFile(distFS, name)
	if err != nil {
		return nil, false
	}
	return data, true
}

func contentType(name string) string {
	switch path.Ext(name) {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js":
		return "text/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}

func registerRoutesAuthGroup(svc *RouterService, groups *RouterGroups) {
	// node
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/nodes", controllers.NodeController)

	// project
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/projects", controllers.ProjectController)

	// user
	svc.RegisterListActionControllerToGroup(groups.AuthGroup, "/users", controllers.UserController)

	// spider
	svc.RegisterListActionControllerToGroup(groups.AuthGroup, "/spiders", controllers.SpiderController)

	// task
	svc.RegisterListActionControllerToGroup(groups.AuthGroup, "/tasks", controllers.TaskController)

	// tag
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/tags", controllers.TagController)

	// setting
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/settings", controllers.SettingController)

	// color
	svc.RegisterActionControllerToGroup(groups.AuthGroup, "/colors", controllers.ColorController)

	// plugin
	svc.RegisterListActionControllerToGroup(groups.AuthGroup, "/plugins", controllers.PluginController)

	// data collection
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/data/collections", controllers.DataCollectionController)

	// result
	svc.RegisterActionControllerToGroup(groups.AuthGroup, "/results", controllers.ResultController)

	// schedule
	svc.RegisterListActionControllerToGroup(groups.AuthGroup, "/schedules", controllers.ScheduleController)

	// stats
	svc.RegisterActionControllerToGroup(groups.AuthGroup, "/stats", controllers.StatsController)

	// token
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/tokens", controllers.TokenController)

	// plugin do
	svc.RegisterActionControllerToGroup(groups.AuthGroup, "/plugin-proxy", controllers.PluginProxyController)

	// git
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/gits", controllers.GitController)

	// role
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/roles", controllers.RoleController)

	// permission
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/permissions", controllers.PermissionController)

	// export
	svc.RegisterActionControllerToGroup(groups.AuthGroup, "/export", controllers.ExportController)

	// env deps
	svc.RegisterActionControllerToGroup(groups.AuthGroup, "/env/deps", controllers.EnvDepsController)

	// notification
	svc.RegisterActionControllerToGroup(groups.AuthGroup, "/notifications", controllers.NotificationController)

	// filter
	svc.RegisterActionControllerToGroup(groups.AuthGroup, "/filters", controllers.FilterController)

	// data sources
	svc.RegisterListActionControllerToGroup(groups.AuthGroup, "/data-sources", controllers.DataSourceController)

	// environments
	svc.RegisterListActionControllerToGroup(groups.AuthGroup, "/environments", controllers.EnvironmentController)
}

func registerRoutesFilterGroup(svc *RouterService, groups *RouterGroups) {
	// filer
	svc.RegisterActionControllerToGroup(groups.FilerGroup, "", controllers.FilerController)
}
