package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

var TeamColour = map[string]string{
	"Red Bull Racing":   "\033[38;2;54;113;198m",
	"Mercedes":          "\033[38;2;0;210;190m",
	"Ferrari":           "\033[38;2;220;0;0m",
	"McLaren":           "\033[38;2;255;165;0m",
	"Aston Martin":      "\033[38;2;0;115;207m",
	"Alpine":            "\033[38;2;0;142;210m",
	"AlphaTauri":        "\033[38;2;39;55;139m",
	"Alfa Romeo Racing": "\033[38;2;138;0;0m",
	"Haas F1 Team":      "\033[38;2;230;41;55m",
	"Williams":          "\033[38;2;0;79;158m",
}

type DriverStanding struct {
	Position   string `json:"position"`
	DriverName string `json:"driver_name"`
	Points     string `json:"points"`
}

type Driver struct {
	BroadcastName string `json:"broadcast_name"`
	CountryCode   string `json:"country_code"`
	DriverNumber  int    `json:"driver_number"`
	FirstName     string `json:"first_name"`
	FullName      string `json:"full_name"`
	HeadshotURL   string `json:"headshot_url"`
	LastName      string `json:"last_name"`
	MeetingKey    int    `json:"meeting_key"`
	NameAcronym   string `json:"name_acronym"`
	SessionKey    int    `json:"session_key"`
	TeamColour    string `json:"team_colour"`
	TeamName      string `json:"team_name"`
}

type Car struct {
	Brake        int     `json:"brake"`
	Date         string  `json:"date"`
	DriverNumber int     `json:"driver_number"`
	DRS          int     `json:"drs"`
	MeetingKey   int     `json:"meeting_key"`
	NGear        int     `json:"n_gear"`
	RPM          int     `json:"rpm"`
	SessionKey   int     `json:"session_key"`
	Speed        float64 `json:"speed"`
	Throttle     int     `json:"throttle"`
}

func asciiImage(img string) (string, error) {
	cmd := exec.Command("jp2a", img, "--color", "--width=45")
	output, err := cmd.Output()

	if err != nil {
		return "", fmt.Errorf("Error converting image to ASCII: %v", err)
	}

	return string(output), nil
}

func driverLeaderboard(cmd *cobra.Command, args []string) {
	client := resty.New()
	resp, err := client.R().
		SetResult(&[]DriverStanding{}).
		Get("https://api.openf1.org/v1/")

	if err != nil {
		log.Fatalf("Error fetching standings: %v", err)
	}

	standings := *resp.Result().(*[]DriverStanding)

	for _, standing := range standings {
		fmt.Printf("%s: %s [%s points]", standing.Position, standing.DriverName, standing.Points)
	}
}

func getDriverDetails(lastName string) (*Driver, error) {
	url := fmt.Sprintf("https://api.openf1.org/v1/drivers?last_name=%s", lastName)

	c := resty.New()
	resp, err := c.R().
		SetResult(&[]Driver{}).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("Error fetching driver details: %v", err)
	}

	driver := *resp.Result().(*[]Driver)

	if len(driver) == 0 {
		return nil, fmt.Errorf("Driver not found")
	}

	return &driver[0], nil
}

func displayDriver(cmd *cobra.Command, args []string) {
	lastName := args[0]

	driver, err := getDriverDetails(lastName)

	if err != nil {
		log.Fatal("Error fecthing driver details: ", err)
	}

	driverIcon, err := asciiImage(driver.HeadshotURL)

	fmt.Println(driverIcon)
	fmt.Printf("%s	 %s %s [%d] // %s\n", TeamColour[driver.TeamName], driver.FirstName, driver.LastName, driver.DriverNumber, driver.TeamName)
}

func displaySpeed(cmd *cobra.Command, args []string) {
	driverNumber := args[0]
	url := fmt.Sprintf("https://api.openf1.org/v1/car_data?driver_number=%s&session_key=latest", driverNumber)

	c := resty.New()
	resp, _ := c.R().
		SetResult(&[]Car{}).
		Get(url)

	car := *resp.Result().(*[]Car)

	fmt.Printf("Speed: %.2f km/h\n", car[0].Speed)
}

func main() {
	var rootCmd = &cobra.Command{Use: "f1"}

	var standingsCmd = &cobra.Command{
		Use:   "standings",
		Short: "Get current F1 driver standings",
		Run:   driverLeaderboard,
	}

	var driverCmd = &cobra.Command{
		Use:   "driver [last name]",
		Short: "Display driver info",
		Run:   displayDriver,
	}

	var speedCmd = &cobra.Command{
		Use:   "speed [driver number]",
		Short: "Get speed of driver",
		Run:   displaySpeed,
	}

	rootCmd.AddCommand(standingsCmd)
	rootCmd.AddCommand(driverCmd)
	rootCmd.AddCommand(speedCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
