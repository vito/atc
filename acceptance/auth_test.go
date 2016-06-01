package acceptance_test

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lib/pq"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/ginkgomon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"

	"github.com/concourse/atc"
	"github.com/concourse/atc/db"
)

var _ = Describe("Auth", func() {
	var atcProcess ifrit.Process
	var dbListener *pq.Listener
	var atcPort uint16
	var teamDB db.TeamDB

	BeforeEach(func() {
		postgresRunner.Truncate()
		dbConn = db.Wrap(postgresRunner.Open())
		dbListener = pq.NewListener(postgresRunner.DataSourceName(), time.Second, time.Minute, nil)
		bus := db.NewNotificationsBus(dbListener, dbConn)
		sqlDB = db.NewSQL(dbConn, bus)

		err := sqlDB.DeleteTeamByName(atc.DefaultPipelineName)
		Expect(err).NotTo(HaveOccurred())
		_, err = sqlDB.CreateTeam(db.Team{Name: atc.DefaultTeamName})
		Expect(err).NotTo(HaveOccurred())

		teamDBFactory := db.NewTeamDBFactory(dbConn)
		teamDB = teamDBFactory.GetTeamDB(atc.DefaultTeamName)

		_, _, err = teamDB.SaveConfig(atc.DefaultPipelineName, atc.Config{}, db.ConfigVersion(1), db.PipelineUnpaused)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		ginkgomon.Interrupt(atcProcess)

		Expect(dbConn.Close()).To(Succeed())
		Expect(dbListener.Close()).To(Succeed())
	})

	Describe("GitHub Auth", func() {
		BeforeEach(func() {
			atcProcess, atcPort, _ = startATC(atcBin, 1, false, []string{}, GITHUB_AUTH)
		})

		It("forces a redirect to /teams/main/login", func() {
			request, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/", atcPort), nil)
			resp, err := http.DefaultClient.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Request.URL.Path).To(Equal("/teams/main/login"))

			team, _, err := teamDB.GetTeam()
			Expect(err).NotTo(HaveOccurred())
			Expect(team.GitHubAuth.ClientID).To(Equal("admin"))
			Expect(team.GitHubAuth.ClientSecret).To(Equal("password"))
			Expect(team.GitHubAuth.Organizations).To(Equal([]string{"myorg"}))
		})
	})

	Describe("GitHub Enterprise Auth", func() {
		BeforeEach(func() {
			atcProcess, atcPort, _ = startATC(atcBin, 1, false, []string{}, GITHUB_ENTERPRISE_AUTH)
		})

		It("forces a redirect to override github", func() {
			request, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/auth/github?redirect=%2F", atcPort), nil)

			client := new(http.Client)
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return errors.New("error")
			}
			resp, err := client.Do(request)
			Expect(err.Error()).To(ContainSubstring("https://github.example.com/login/oauth/authorize"))
			Expect(resp.StatusCode).To(Equal(http.StatusTemporaryRedirect))
		})
	})

	Describe("Basic Auth", func() {
		BeforeEach(func() {
			atcProcess, atcPort, _ = startATC(atcBin, 1, false, []string{}, BASIC_AUTH)
		})

		It("forces a redirect to /login", func() {
			request, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/", atcPort), nil)
			resp, err := http.DefaultClient.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Request.URL.Path).To(Equal("/teams/main/login"))
		})

		It("logs in with Basic Auth and allows access", func() {
			request, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/", atcPort), nil)
			request.SetBasicAuth("admin", "password")
			resp, err := http.DefaultClient.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Request.URL.Path).To(Equal("/"))
		})
	})

	Context("when basic auth is misconfigured", func() {
		It("errors when only username is specified", func() {
			atcCommand, _, _ := getATCCommand(atcBin, 1, false, []string{}, BASIC_AUTH_NO_PASSWORD)
			session, err := gexec.Start(atcCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("must specify --basic-auth-password to use basic auth"))
		})

		It("errors when only password is specified", func() {
			atcCommand, _, _ := getATCCommand(atcBin, 1, false, []string{}, BASIC_AUTH_NO_USERNAME)
			session, err := gexec.Start(atcCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("must specify --basic-auth-username to use basic auth"))
		})
	})

	Describe("No authentication via development mode", func() {
		BeforeEach(func() {
			atcProcess, atcPort, _ = startATC(atcBin, 1, false, []string{}, DEVELOPMENT_MODE)
		})

		It("logs in without authentication", func() {
			request, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/", atcPort), nil)
			resp, err := http.DefaultClient.Do(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Request.URL.Path).To(Equal("/"))
		})
	})

	Describe("when auth is not configured", func() {
		It("returns an error", func() {
			atcCommand, _, _ := getATCCommand(atcBin, 1, false, []string{}, NOT_CONFIGURED_AUTH)
			session, err := gexec.Start(atcCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("must configure basic auth, OAuth, or turn on development mode"))
		})
	})

	Describe("CF Auth", func() {
		BeforeEach(func() {
			atcProcess, atcPort, _ = startATC(atcBin, 1, false, []string{}, CF_AUTH)
		})

		It("forces a redirect to CF auth URL", func() {
			request, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/auth/cf?redirect=%2F", atcPort), nil)

			client := new(http.Client)
			client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return errors.New("error")
			}
			resp, err := client.Do(request)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("https://uaa.example.com/oauth/authorize"))
			Expect(resp.StatusCode).To(Equal(http.StatusTemporaryRedirect))
		})

		It("requires client id and client secret to be specified", func() {
			atcCommand, _, _ := getATCCommand(atcBin, 1, false, []string{}, CF_AUTH_NO_CLIENT_SECRET)
			session, err := gexec.Start(atcCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("must specify --cf-auth-client-id and --cf-auth-client-secret to use CF OAuth"))
		})

		It("requires space guid to be specified", func() {
			atcCommand, _, _ := getATCCommand(atcBin, 1, false, []string{}, CF_AUTH_NO_SPACE)
			session, err := gexec.Start(atcCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("must specify --cf-auth-space to use CF OAuth"))
		})

		It("requires auth, token and api url to be specified", func() {
			atcCommand, _, _ := getATCCommand(atcBin, 1, false, []string{}, CF_AUTH_NO_TOKEN_URL)
			session, err := gexec.Start(atcCommand, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say("must specify --cf-auth-auth-url, --cf-auth-token-url and --cf-auth-api-url to use CF OAuth"))
		})
	})
})
