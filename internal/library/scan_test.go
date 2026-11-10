package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseEpisodeName(t *testing.T) {
	cases := []struct {
		in    string
		s, e  int
		title string
	}{
		{"Episode S01E03 Ireland", 1, 3, "Ireland"},
		{"Episode S01E01_Serenity", 1, 1, "Serenity"},
		{"Episode S01E01 - Carovna trieda", 1, 1, "Carovna trieda"},
		{"Episode S01E03", 1, 3, ""},
		{"Show.S10E07.720p", 10, 7, "720p"},
		{"s01e91 Feature 1", 1, 91, "Feature 1"},
		{"1x03 - Name", 1, 3, "Name"},
		{"Show S01E01-E02 Double", 1, 1, "Double"},
		{"05 Something", -1, 5, "Something"},
		{"Just a title", -1, -1, "Just a title"},
	}
	for _, c := range cases {
		s, e, title := parseEpisodeName(c.in)
		if s != c.s || e != c.e || title != c.title {
			t.Errorf("parseEpisodeName(%q) = %d, %d, %q; want %d, %d, %q", c.in, s, e, title, c.s, c.e, c.title)
		}
	}
}

func TestSortName(t *testing.T) {
	for in, want := range map[string]string{"The Fifth Element": "fifth element", "A Clockwork Orange": "clockwork orange",
		"Řád Ohně": "rad ohne", "300": "300", "The": "the"} {
		if got := SortName(in); got != want {
			t.Errorf("SortName(%q) = %q, want %q", in, got, want)
		}
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const movieNFO = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<!--created by tinyMediaManager-->
<movie>
  <title>300</title>
  <year>2007</year>
  <ratings><rating default="false" max="10" name="themoviedb"><value>7.2</value><votes>12199</votes></rating></ratings>
  <plot>Spartans.</plot>
  <tagline>Spartans, prepare for glory!</tagline>
  <runtime>117</runtime>
  <mpaa>US:R / US:Rated R</mpaa>
  <uniqueid default="false" type="tmdb">1271</uniqueid>
  <uniqueid default="true" type="imdb">tt0416449</uniqueid>
  <premiered>2007-03-07</premiered>
  <genre>Action</genre><genre>War</genre>
  <studio>Warner Bros. Pictures</studio>
  <credits tmdbid="2293">Frank Miller</credits>
  <director tmdbid="15217">Zack Snyder</director>
  <actor><name>Gerard Butler</name><role>King Leonidas</role><thumb>https://image.tmdb.org/x.jpg</thumb></actor>
  <set><name>300 Collection</name><overview/></set>
  <dateadded>2023-01-02 18:09:47</dateadded>
</movie>`

func TestScan(t *testing.T) {
	root := t.TempDir()
	mv, tv := filepath.Join(root, "movies"), filepath.Join(root, "tv")
	write(t, filepath.Join(mv, "300", "300.m4v"), "v")
	write(t, filepath.Join(mv, "300", "300 - Cesky.m4v"), "v")
	write(t, filepath.Join(mv, "300", "300.nfo"), movieNFO)
	write(t, filepath.Join(mv, "300", "300-poster.jpg"), "img")
	write(t, filepath.Join(mv, "300", "300-fanart.jpg"), "img")
	write(t, filepath.Join(mv, "300", "._300.m4v"), "appledouble")
	write(t, filepath.Join(mv, "300", "extras", "Making of.m4v"), "v")
	write(t, filepath.Join(mv, "300", "300.cs.forced.srt"), "1")
	write(t, filepath.Join(mv, "No NFO (1999)", "No NFO (1999).mp4"), "v")
	write(t, filepath.Join(mv, "LOTR", "LOTR.nfo"), "<movie><title>The Two Towers</title></movie>")
	write(t, filepath.Join(mv, "LOTR", "LOTR - Disc2.m4v"), "v")
	write(t, filepath.Join(mv, "LOTR", "LOTR - Disc1.m4v"), "v")
	write(t, filepath.Join(mv, "Single", "Single - Part1.m4v"), "v")
	write(t, filepath.Join(mv, "No NFO (1999)", "poster.jpg"), "img")

	write(t, filepath.Join(tv, "Firefly", "tvshow.nfo"), `<tvshow><title>Firefly</title><status>Ended</status>
		<namedseason number="1">The Only Season</namedseason><uniqueid type="tvdb">78874</uniqueid></tvshow>`)
	write(t, filepath.Join(tv, "Firefly", "poster.jpg"), "img")
	write(t, filepath.Join(tv, "Firefly", "season01-poster.jpg"), "img")
	write(t, filepath.Join(tv, "Firefly", "Season 01", "Episode S01E02_The_Train_Job.m4v"), "v")
	write(t, filepath.Join(tv, "Firefly", "Season 01", "Episode S01E01_Serenity.m4v"), "v")
	write(t, filepath.Join(tv, "Firefly", "Season 01", "Episode S01E01_Serenity.nfo"),
		`<episodedetails><title>Serenity</title><season>1</season><episode>1</episode><aired>2002-12-20</aired></episodedetails>`)
	write(t, filepath.Join(tv, "Firefly", "Season 01", "Episode S01E01_Serenity-thumb.jpg"), "img")
	write(t, filepath.Join(tv, "Firefly", "Season 01", "extras", "Docs.m4v"), "v")
	write(t, filepath.Join(tv, "Bien", "Season 94 2021 From Redux", "Episode S94E01 Intro.m4v"), "v")
	write(t, filepath.Join(tv, "Loose", "S02E05 Five.mkv"), "v")
	write(t, filepath.Join(tv, "Empty", "tvshow.nfo"), "<tvshow><title>Empty</title></tvshow>")

	lib, _ := (&Scanner{Libraries: []LibraryConfig{
		{Key: "movies", Name: "Movies", CollectionType: "movies", Path: mv},
		{Key: "tv", Name: "TV", CollectionType: "tvshows", Path: tv},
	}}).Scan()

	movies := lib.Views[0].Children
	if len(movies) != 4 {
		t.Fatalf("movies: got %d, want 4", len(movies))
	}
	byName := map[string]*Item{}
	for _, m := range movies {
		byName[m.Name] = m
	}
	if tt := byName["The Two Towers"]; tt == nil || len(tt.Parts) != 1 || filepath.Base(tt.Path) != "LOTR - Disc1.m4v" ||
		filepath.Base(tt.Parts[0].Path) != "LOTR - Disc2.m4v" || lib.Items[tt.Parts[0].ID] != tt.Parts[0] {
		t.Errorf("stacked movie: %+v", tt)
	}
	if byName["Single - Part1"] == nil {
		t.Errorf("lone part should stay a plain movie: %v", byName)
	}
	movies = []*Item{byName["300"], byName["No NFO"]}
	m := movies[0] // "300" sorts before "no nfo"
	if m.Name != "300" || m.Year != 2007 || m.OfficialRating != "R" || m.CommunityRating != 7.2 {
		t.Errorf("300 metadata: %+v", m)
	}
	if m.ProviderIDs["Imdb"] != "tt0416449" || m.ProviderIDs["Tmdb"] != "1271" {
		t.Errorf("provider ids: %v", m.ProviderIDs)
	}
	if len(m.Sources) != 2 || m.Sources[1].Name != "Cesky" || m.Sources[0].ID != m.ID {
		t.Errorf("versions: %+v", m.Sources)
	}
	if len(m.Sources[0].Subtitles) != 1 || !m.Sources[0].Subtitles[0].Forced || m.Sources[0].Subtitles[0].Language != "cs" {
		t.Errorf("subtitles: %+v", m.Sources[0].Subtitles)
	}
	if len(m.Extras) != 1 || m.Extras[0].Name != "Making of" {
		t.Errorf("extras: %+v", m.Extras)
	}
	if len(m.Images[ImgPrimary]) != 1 || len(m.Images[ImgBackdrop]) != 1 {
		t.Errorf("images: %+v", m.Images)
	}
	if m.RuntimeTicks != 117*60*1e7 || m.SetName != "300 Collection" || len(m.People) != 3 {
		t.Errorf("runtime/set/people: %d %q %d", m.RuntimeTicks, m.SetName, len(m.People))
	}
	other := movies[0]
	if other == m {
		other = movies[1]
	}
	if other.Name != "No NFO" || other.Year != 1999 || len(other.Images[ImgPrimary]) != 1 {
		t.Errorf("no-nfo movie: %+v", other)
	}

	shows := lib.Views[1].Children
	names := map[string]*Item{}
	for _, s := range shows {
		names[s.Name] = s
	}
	if len(shows) != 3 || names["Empty"] != nil {
		t.Fatalf("shows: got %d (%v)", len(shows), names)
	}
	ff := names["Firefly"]
	if len(ff.Children) != 1 || ff.Children[0].Name != "The Only Season" || len(ff.Children[0].Images[ImgPrimary]) != 1 {
		t.Fatalf("firefly seasons: %+v", ff.Children)
	}
	eps := ff.Children[0].Children
	if len(eps) != 2 || eps[0].Name != "Serenity" || eps[1].Name != "The Train Job" || eps[1].Index != 2 {
		t.Errorf("firefly episodes: %q %q", eps[0].Name, eps[1].Name)
	}
	if len(eps[0].Images[ImgPrimary]) != 1 || eps[0].Premiere.Year() != 2002 {
		t.Errorf("episode thumb/aired: %+v", eps[0])
	}
	if len(ff.Children[0].Extras) != 1 {
		t.Errorf("season extras: %+v", ff.Children[0].Extras)
	}
	if b := names["Bien"]; b.Children[0].Index != 94 || b.Children[0].Name != "Season 94 2021 From Redux" {
		t.Errorf("bien season: %+v", b.Children[0])
	}
	if l := names["Loose"]; l.Children[0].Index != 2 || l.Children[0].Children[0].Index != 5 {
		t.Errorf("loose episode: %+v", l.Children[0])
	}
	if got := lib.Items[eps[0].ID]; got != eps[0] || eps[0].Series() != ff {
		t.Errorf("index/series links broken")
	}
	if NormalizeID("9DD4E461-268C-8034-F5C8-564E155C67A6") != "9dd4e461268c8034f5c8564e155c67a6" ||
		NormalizeID("person-c8c1e985132cd8a8b3796aefa8f35cc5") != "c8c1e985132cd8a8b3796aefa8f35cc5" {
		t.Errorf("NormalizeID")
	}
}
