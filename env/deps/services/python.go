package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os/exec"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	entity2 "github.com/crawlab-team/crawlab-core/entity"
	"github.com/crawlab-team/crawlab-core/env/deps/constants"
	"github.com/crawlab-team/crawlab-core/env/deps/entity"
	"github.com/crawlab-team/crawlab-core/env/deps/models"
	"github.com/crawlab-team/go-trace"
	"github.com/imroc/req"
	"go.mongodb.org/mongo-driver/bson"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
)

type PythonService struct {
	*BaseLangService
}

func (svc *PythonService) Init() {
}

func (svc *PythonService) GetRepoList(query string, pagination *entity2.Pagination) (deps []models.Dependency, total int, err error) {
	if strings.TrimSpace(query) == "" {
		return nil, 0, nil
	}

	depNames, total, err := svc.searchPackageNames(query, pagination)
	if err != nil {
		return nil, 0, trace.TraceError(err)
	}
	if total == 0 {
		return nil, 0, nil
	}

	deps = make([]models.Dependency, len(depNames))
	for i, depName := range depNames {
		deps[i] = models.Dependency{Name: depName}
	}
	svc.populateLatestVersions(deps)

	depsResults, err := svc.getDependencyResults(depNames)
	if err != nil {
		return nil, 0, trace.TraceError(err)
	}

	depsResultsMap := map[string]entity.DependencyResult{}
	for _, dr := range depsResults {
		depsResultsMap[dr.Name] = dr
	}

	for i, d := range deps {
		dr, ok := depsResultsMap[d.Name]
		if ok {
			deps[i].Result = dr
		}
	}

	return deps, total, nil
}

func (svc *PythonService) searchPackageNames(query string, pagination *entity2.Pagination) (depNames []string, total int, err error) {
	reqSession := req.New()
	reqSession.SetTimeout(15 * time.Second)

	res, err := reqSession.Get("https://pypi.org/simple/", req.Header{"Accept": "application/vnd.pypi.simple.v1+json"})
	if err != nil {
		if res != nil {
			err = errors.New(res.String())
			return nil, 0, trace.TraceError(err)
		}
		return nil, 0, trace.TraceError(err)
	}

	var simpleRes struct {
		Projects []struct {
			Name string `json:"name"`
		} `json:"projects"`
	}
	if err := res.ToJSON(&simpleRes); err != nil {
		return nil, 0, trace.TraceError(err)
	}

	normalizedQuery := normalizePythonPackageName(query)
	for _, p := range simpleRes.Projects {
		if strings.Contains(normalizePythonPackageName(p.Name), normalizedQuery) {
			depNames = append(depNames, p.Name)
		}
	}
	sort.Strings(depNames)

	total = len(depNames)
	page := pagination.Page
	if page <= 0 {
		page = 1
	}
	size := pagination.Size
	if size <= 0 {
		size = 10
	}
	start := (page - 1) * size
	if start >= total {
		return nil, total, nil
	}
	end := min(start+size, total)
	return depNames[start:end], total, nil
}

func (svc *PythonService) populateLatestVersions(deps []models.Dependency) {
	wg := sync.WaitGroup{}
	wg.Add(len(deps))
	for i := range deps {
		go func(i int) {
			defer wg.Done()
			v, err := svc.GetLatestVersion(deps[i])
			if err == nil {
				deps[i].LatestVersion = v
			}
		}(i)
	}
	wg.Wait()
}

func (svc *PythonService) getDependencyResults(depNames []string) (depsResults []entity.DependencyResult, err error) {
	pipelines := mongo2.Pipeline{
		{{
			"$match",
			bson.M{
				"type": constants.DependencyTypePython,
				"name": bson.M{
					"$in": depNames,
				},
			},
		}},
		{{
			"$group",
			bson.M{
				"_id": "$name",
				"node_ids": bson.M{
					"$push": "$node_id",
				},
				"versions": bson.M{
					"$addToSet": "$version",
				},
			},
		}},
		{{
			"$project",
			bson.M{
				"name":     "$_id",
				"node_ids": "$node_ids",
				"versions": "$versions",
			},
		}},
	}
	if err := svc.parent.colD.Aggregate(pipelines, nil).All(&depsResults); err != nil {
		return nil, trace.TraceError(err)
	}
	return depsResults, nil
}

func normalizePythonPackageName(name string) string {
	return strings.NewReplacer("_", "-", ".", "-").Replace(strings.ToLower(name))
}

func (svc *PythonService) GetDependencies(params entity.UpdateParams) (deps []models.Dependency, err error) {
	cmd := exec.Command(params.Cmd, "list", "--format", "json")
	data, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var _deps []models.Dependency
	if err := json.Unmarshal(data, &_deps); err != nil {
		return nil, err
	}
	for _, d := range _deps {
		if strings.HasPrefix(d.Name, "-") {
			continue
		}
		d.Type = constants.DependencyTypePython
		deps = append(deps, d)
	}
	return deps, nil
}

func (svc *PythonService) InstallDependencies(params entity.InstallParams) (err error) {
	// arguments
	var args []string

	// install
	args = append(args, "install")

	// proxy
	if params.Proxy != "" {
		args = append(args, "-i")
		args = append(args, params.Proxy)
	}

	if params.UseConfig {
		// workspace path
		workspacePath, err := svc._getInstallWorkspacePath(params)
		if err != nil {
			return err
		}

		// config path
		configPath := path.Join(workspacePath, constants.DependencyConfigRequirementsTxt)

		// use config
		args = append(args, "-r")
		args = append(args, configPath)
	} else {
		// upgrade
		if params.Upgrade {
			args = append(args, "-U")
		}

		// dependency names
		for _, depName := range params.Names {
			args = append(args, depName)
		}
	}

	// command
	cmd := exec.Command(params.Cmd, args...)

	// logging
	svc.parent._configureLogging(params.TaskId, cmd)

	// start
	if err := cmd.Start(); err != nil {
		return trace.TraceError(err)
	}

	// wait
	if err := cmd.Wait(); err != nil {
		return trace.TraceError(err)
	}

	return nil
}

func (svc *PythonService) UninstallDependencies(params entity.UninstallParams) (err error) {
	// arguments
	var args []string

	// uninstall
	args = append(args, "uninstall")
	args = append(args, "-y")

	// dependency names
	for _, depName := range params.Names {
		args = append(args, depName)
	}

	// command
	cmd := exec.Command(params.Cmd, args...)

	// logging
	svc.parent._configureLogging(params.TaskId, cmd)

	// start
	if err := cmd.Start(); err != nil {
		return trace.TraceError(err)
	}

	// wait
	if err := cmd.Wait(); err != nil {
		return trace.TraceError(err)
	}

	return nil
}

func (svc *PythonService) GetLatestVersion(dep models.Dependency) (v string, err error) {
	reqSession := req.New()
	reqSession.SetTimeout(60 * time.Second)

	requestUrl := fmt.Sprintf("https://pypi.org/pypi/%s/json", url.PathEscape(dep.Name))
	res, err := reqSession.Get(requestUrl)
	if err != nil {
		return "", trace.TraceError(err)
	}

	var pypiRes struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	if err := res.ToJSON(&pypiRes); err != nil {
		return "", trace.TraceError(err)
	}

	return strings.TrimSpace(pypiRes.Info.Version), nil
}

func NewPythonService(parent *Service) (svc *PythonService) {
	svc = &PythonService{}
	baseSvc := newBaseService(
		svc,
		parent,
		constants.DependencyTypePython,
		entity.MessageCodes{
			Update:    constants.MessageCodePythonUpdate,
			Save:      constants.MessageCodePythonSave,
			Install:   constants.MessageCodePythonInstall,
			Uninstall: constants.MessageCodePythonUninstall,
		},
	)
	svc.BaseLangService = baseSvc
	return svc
}
