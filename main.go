package main

import (
	"context"
	"fmt"
	"sort"
	"os"
	"bufio"
	"strings"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-github/github"
)

type InputRepo struct {
	owner string
	repo string
	min string
}

func sameRelease(r1, r2 *semver.Version) bool {
	preMajor := r1.Major
	preMinor := r1.Minor
	return (preMajor == r2.Major && preMinor == r2.Minor)
}

// LatestVersions returns a sorted slice with the highest version as its first element and the highest version of the smaller minor versions in a descending order
func LatestVersions(releases []*semver.Version, minVersion *semver.Version) []*semver.Version {
	var versionSlice []*semver.Version
	var result []*semver.Version

	// filter out the old versions
	for _, v := range releases {
		if !v.LessThan(*minVersion) && v.PreRelease == "" {
			versionSlice = append(versionSlice, v)
		}
	}

	// sort the versionSlice in descending order
	sort.Sort(sort.Reverse(semver.Versions(versionSlice)))

	// take for each minor the highest patch for that minor
	if len(versionSlice)>0 {
		result = append(result, versionSlice[0])
		preRelease := versionSlice[0]
		for _, v := range versionSlice{
			if !sameRelease(preRelease, v) {
				result = append(result, v)
				preRelease = v
			}
		}
	}

	return result
}

// Here we implement the basics of communicating with github through the library as well as printing the version
// You will need to implement LatestVersions function as well as make this application support the file format outlined in the README
// Please use the format defined by the fmt.Printf line at the bottom, as we will define a passing coding challenge as one that outputs
// the correct information, including this line
func main() {
	// save the input file for computing
	var input []InputRepo

	// open the input file
	if len(os.Args) <= 1 {
		fmt.Println("Please input a path to an input file!")
		os.Exit(1)
	}

	// open file
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
    }	
    defer f.Close()

    // read file
    scanner := bufio.NewScanner(f)
    scanner.Scan()
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ",")
		repo := strings.Split(line[0], "/")
		input = append(input, InputRepo{repo[0], repo[1], line[1]})
	}

	// Github
	client := github.NewClient(nil)
	ctx := context.Background()
	opt := &github.ListOptions{PerPage: 10}
	for _, inputRepo := range input {
		
		// get releases content
		releases, _, err := client.Repositories.ListReleases(ctx, inputRepo.owner, inputRepo.repo, opt)
		if err != nil {
			// panic(err) // is this really a good way?
			fmt.Println(err)
			fmt.Printf("The repository name %s/%s may be incorrect!!!\n", inputRepo.owner, inputRepo.repo)
			continue
		}

		// extract versions
		minTrimmedSpace := strings.TrimSpace(inputRepo.min)
		minVersion := semver.New(minTrimmedSpace)
		allReleases := make([]*semver.Version, len(releases))
		for i, release := range releases {
			versionString := *release.TagName
			if versionString[0] == 'v' {
				versionString = versionString[1:]
			}
			allReleases[i] = semver.New(versionString)
		}

		versionSlice := LatestVersions(allReleases, minVersion)
		fmt.Printf("latest versions of %s/%s: %s\n", inputRepo.owner, inputRepo.repo, versionSlice)
	}

}
