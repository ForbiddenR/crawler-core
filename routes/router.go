package routes

import (
	"fmt"
	"io"
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

func (svc *RouterService) RegisterActionControllerToGroup(group *gin.RouterGroup, basePath string, ctr controllers.ActionController) {
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
	basePath := ""
	// routes groups
	groups := NewRouterGroups(app)

	// router service
	svc := NewRouterService(app)

	// register routes
	registerRoutesAnonymousGroup(svc, groups)
	registerRoutesAuthGroup(svc, groups)
	registerRoutesFilterGroup(svc, groups)

	if distFS, err := web.DistFS(); err == nil {
		if indexFile, err := distFS.Open("index.html"); err == nil {
			_ = indexFile.Close()
			httpFS := http.FS(distFS)
			serveIndex := func(ctx *gin.Context) {
				indexHTML, err := renderIndexHTML(distFS, basePath)
				if err != nil {
					ctx.Status(http.StatusNotFound)
					return
				}
				setHTMLCacheHeaders(ctx)
				ctx.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
			}
			serveAsset := func(ctx *gin.Context) {
				assetPath := "assets/" + strings.TrimPrefix(ctx.Param("filepath"), "/")
				if assetFile, err := distFS.Open(assetPath); err == nil {
					_ = assetFile.Close()
					setStaticAssetCacheHeaders(ctx)
					ctx.FileFromFS(assetPath, httpFS)
					return
				}
				ctx.Status(http.StatusNotFound)
			}

			app.GET("/", serveIndex)
			app.GET("/assets/*filepath", serveAsset)
			app.HEAD("/assets/*filepath", serveAsset)
			app.NoRoute(func(ctx *gin.Context) {
				requestPath, ok := stripBasePath(basePath, ctx.Request.URL.Path)
				if !ok {
					ctx.Status(http.StatusNotFound)
					return
				}

				if strings.HasPrefix(requestPath, "/api") {
					ctx.Status(http.StatusNotFound)
					return
				}

				if assetPath, ok := staticAssetPath(requestPath); ok {
					if assetFile, err := distFS.Open(assetPath); err == nil {
						_ = assetFile.Close()
						setStaticAssetCacheHeaders(ctx)
						ctx.FileFromFS(assetPath, httpFS)
						return
					}
				}
				serveIndex(ctx)
			})
		}
	}

	return nil
}

func registerRoutesAnonymousGroup(svc *RouterService, groups *RouterGroups) {
	// login
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/", controllers.LoginController)

	// version
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/version", controllers.VersionController)

	// system info
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/system-info", controllers.SystemInfoController)

	// demo
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/demo", controllers.DemoController)

	// sync
	svc.RegisterActionControllerToGroup(groups.AnonymousGroup, "/sync", controllers.SyncController)
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

	// git
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/gits", controllers.GitController)

	// role
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/roles", controllers.RoleController)

	// permission
	svc.RegisterListControllerToGroup(groups.AuthGroup, "/permissions", controllers.PermissionController)

	// export
	svc.RegisterActionControllerToGroup(groups.AuthGroup, "/export", controllers.ExportController)

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

func setHTMLCacheHeaders(ctx *gin.Context) {
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("Expires", "0")
	ctx.Header("Pragma", "no-cache")
}

func setStaticAssetCacheHeaders(c *gin.Context) {
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
}

func renderIndexHTML(staticFS fs.FS, basePath string) ([]byte, error) {
	indexFile, err := staticFS.Open("index.html")
	if err != nil {
		return nil, err
	}
	defer indexFile.Close()
	return io.ReadAll(indexFile)
}

func cleanURLPath(requestPath string) string {
	cleaned := path.Clean(requestPath)
	if cleaned == "." {
		return "/"
	}
	if !strings.HasPrefix(cleaned, "/") {
		return "/" + cleaned
	}
	return cleaned
}

func staticAssetPath(requestPath string) (string, bool) {
	cleaned := cleanURLPath(requestPath)
	if strings.Contains(cleaned, "\\") {
		return "", false
	}
	relPath := strings.TrimPrefix(cleaned, "/")
	if relPath == "" {
		return "", false
	}
	return relPath, true
}

func stripBasePath(basePath, requestPath string) (string, bool) {
	cleaned := cleanURLPath(requestPath)
	if basePath == "" {
		return cleaned, true
	}
	if cleaned == basePath {
		return "/", true
	}
	if !strings.HasPrefix(cleaned, basePath+"/") {
		return "", false
	}
	trimmed := strings.TrimPrefix(cleaned, basePath)
	if trimmed == "" {
		return "/", true
	}
	return trimmed, true
}
