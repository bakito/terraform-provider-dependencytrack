package provider_test

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	dtrack "github.com/DependencyTrack/client-go"
	"github.com/bakito/terraform-provider-dependencytrack/internal/provider"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	testExistingUUID = "88888888-8888-8888-8888-888888888888"
	testUUID         = "99999999-9999-9999-9999-999999999999"

	providerConfig = `
provider "dependencytrack" {
  token = "foo"
  host  = "%s"
}
`
)

var (
	// accProtoV6ProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"dependencytrack": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)

// server return a test server and the matching provider config.
func testServer() (*httptest.Server, string) {
	repos := make(map[string]dtrack.Repository)
	testRepo := dtrack.Repository{
		Type:            dtrack.RepositoryTypeGoModules,
		Identifier:      "proxy.golang.org",
		Url:             "https://proxy.golang.org",
		ResolutionOrder: 1,
		Enabled:         true,
		Internal:        false,
		UUID:            uuid.MustParse(testExistingUUID),
	}
	repos[testRepo.UUID.String()] = testRepo

	router := http.NewServeMux()
	router.HandleFunc("/api/v1/repository", serveResponse(repos))
	router.HandleFunc("/api/v1/repository/", serveResponse(repos))
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Printf("Missed path %q in test server!\n", request.RequestURI)
	})

	svr := httptest.NewServer(router)
	return svr, fmt.Sprintf(providerConfig, svr.URL)
}

func serveResponse(repos map[string]dtrack.Repository) func(writer http.ResponseWriter, request *http.Request) {

	return func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == "GET" {
			var repoList []dtrack.Repository
			for _, r := range repos {
				repoList = append(repoList, r)
			}
			b, _ := json.Marshal(&repoList)
			writer.Header().Set("X-Total-Count", fmt.Sprintf("%d", len(repos)))
			_, _ = writer.Write(b)
		} else if request.Method == "DELETE" {
			path := strings.Split(request.RequestURI, "/")
			delete(repos, path[len(path)-1])
		} else {
			defer func() { _ = request.Body.Close() }()
			b, _ := io.ReadAll(request.Body)
			repo := dtrack.Repository{}
			_ = json.Unmarshal(b, &repo)

			if request.Method == "PUT" {
				repo.UUID = uuid.MustParse(testUUID)
				repo.ResolutionOrder = len(repos)
				repo.Internal = false
			} else if r, ok := repos[repo.UUID.String()]; ok {
				repo.ResolutionOrder = r.ResolutionOrder
			}

			repos[repo.UUID.String()] = repo
			b, _ = json.Marshal(&repo)
			_, _ = writer.Write(b)
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
	}
}
