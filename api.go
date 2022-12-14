package main

import (
	"google.golang.org/protobuf/proto"
	"k8s.io/klog/v2"
	"strings"
)

func reposResponse(filter string) []byte {
	reposFromKube := loadReposFromKube()

	repos := &ReposResponse{}

	for nameRepo, urlRepo := range reposFromKube {
		repo := &Repo{}
		if "" == filter || strings.Contains(urlRepo, filter) || strings.Contains(nameRepo, filter) {
			repo.Name = nameRepo
			repo.Url = urlRepo
			repos.Repos = append(repos.Repos, repo)
		}
	}

	marshal, err := proto.Marshal(repos)

	if err != nil {
		klog.Infof("Marshal error: %s", err)
	}
	return marshal
}

func getRepos(message []byte) []byte {
	req := &GetReposRequest{}
	proto.Unmarshal(message, req)

	filter := req.GetFilter()

	return reposResponse(filter)
}

func addRepo(message []byte) []byte {
	req := &AddRepoRequest{}
	proto.Unmarshal(message, req)

	repos := loadReposFromKube()
	repos[req.Name] = req.Url

	saveReposToKube(repos)

	res := &OperationStatus{Status: true}

	marshal, err := proto.Marshal(res)

	if err != nil {
		klog.Infof("Marshal error: %s", err)
	}
	return marshal
}

func cacheConversion(filter string, cached map[string]*map[string][]HelmChart) []*Chart {
	var result = []*Chart{}
	for repoName, charts := range cached {
		for _, chartArray := range *charts {
			chartRest := chartArray[0]

			layout := "2006-01-02T15:04:05-0700"
			if strings.Contains(chartRest.Name, filter) {
				ch := Chart{
					Name:        chartRest.Name,
					Version:     chartRest.Version,
					Repo:        repoName,
					Description: chartRest.Description,
					Icon:        chartRest.Icon,
					Created:     chartRest.Created.Format(layout),
					Digest:      chartRest.Digest,
				}
				result = append(result, &ch)
			}

		}
	}

	return result
}

func getCharts(message []byte) []byte {
	req := &GetChartsRequest{}
	proto.Unmarshal(message, req)
	repos := loadReposFromKube()

	cached := cachedReq(req.Reload, repos)

	response := &ChartsResponse{}

	conversion := cacheConversion(req.Filter, cached)

	response.Charts = conversion

	marshal, err := proto.Marshal(response)

	if err != nil {
		klog.Infof("Marshal error: %s", err)
	}
	return marshal

}

func removeRepo(message []byte) []byte {
	req := &RemoveRepoRequest{}
	proto.Unmarshal(message, req)

	repos := loadReposFromKube()
	klog.Infof("DELETE REPO", req.Name)
	delete(repos, req.Name)
	klog.Infof("MAP TO SAVE", repos)

	saveReposToKube(repos)

	res := &OperationStatus{Status: true}
	marshal, err := proto.Marshal(res)

	if err != nil {
		klog.Infof("Marshal error: %s", err)
	}
	return marshal
}

func processingFunction() func(message []byte, functionId uint8) []byte {
	return func(message []byte, functionId uint8) []byte {
		klog.Infof("Pocessing function %s", functionId)
		if functionId == 1 {
			return getRepos(message)
		}
		if functionId == 2 {
			return addRepo(message)
		}
		if functionId == 3 {
			return removeRepo(message)
		}
		if functionId == 4 {
			return getCharts(message)
		}

		return []byte("FUNCTION_NOT_FOUND")
	}
}
