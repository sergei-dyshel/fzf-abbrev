package abbrev

import (
	"testing"
)

func logMatch(tb testing.TB, input, pattern string) *result {
	res := checkMatch([]rune(input), []rune(pattern))
	if res != nil {
		tb.Logf("pattern: '%s' score: %d input: '%s'", string(pattern),
			res.score, markPosInStr(string(input), res.pos))
	}
	return res
}

func assertMatch(tb testing.TB, input, pattern string) (res *result) {
	res = logMatch(tb, input, pattern)
	if res == nil {
		tb.Fatalf("'%s' should match '%s'", input, pattern)
	}
	return
}

func assertNoMatch(t *testing.T, input, pattern string) (res *result) {
	res = logMatch(t, input, pattern)
	if res != nil {
		t.Fatalf("'%s' should not match '%s'", input, pattern)
	}
	return
}

func assertMatchMany(t *testing.T, input string, patterns ...string) {
	for _, pattern := range patterns {
		assertMatch(t, input, pattern)
	}
}

func assertNoMatchMany(t *testing.T, input string, patterns ...string) {
	for _, pattern := range patterns {
		assertNoMatch(t, input, pattern)
	}
}

func assertBetterMatch(t *testing.T, input1, input2 string, pattern string) {
	res1 := assertMatch(t, input1, pattern)
	res2 := assertMatch(t, input2, pattern)
	if res1.score >= res2.score {
		t.Fatalf("'%s' (score %d) should come before '%s' (score %d) when matching '%s'",
			input1, res1.score, input2, res2.score, pattern)
	}
}

func assertMatchOrder(t *testing.T, pattern string, inputs ...string) {
	for i := 0; i < len(inputs)-1; i++ {
		assertBetterMatch(t, inputs[i], inputs[i+1], pattern)
	}
}

func TestSimple(t *testing.T) {
	Opts.Default()
	assertMatchMany(t, "FooBarQux", "fbq", "foob")
	assertNoMatchMany(t, "FooBarQux", "fbx", "foobr")
	assertMatchMany(t, "foo bar qux", "fbq", "foob")
	assertMatchMany(t, "foo19bar", "f1", "f19", "f1b", "foo1")
}

func TestScoring(t *testing.T) {
	Opts.Default()
	assertMatchOrder(t, "fbq", "FooBarQux", "Foo BarQux", "Foo Bar Qux")
	// assertMatchOrder(t, "For", "ForLoop", ")
}

func TestFilePaths(t *testing.T) {
	Opts.Parse("file-paths")

	assertMatchOrder(t, "fbq", "foo_bar_qux", "foo/bar_qux", "foo/bar/qux")
}

func TestSkippedStart(t *testing.T) {
	Opts.Default()
	assertMatchOrder(t, "fbq", "foo_bar_qux", "some_fbq",
		"some_fb_qux", "some_foo_bar_qux")
}

func benchmarkMatch(b *testing.B, input string, pattern string) {
	inputRunes := []rune(input)
	patternRunes := []rune(pattern)
	for i := 0; i < b.N; i++ {
		checkMatch(inputRunes, patternRunes)
		// fmt.Sprintf("hello")
	}
}

func benchmarkMatchMany(b *testing.B, input string, patterns ...string) {
	for _, pattern := range patterns {
		b.Run(pattern, func(b *testing.B) { benchmarkMatch(b, input, pattern) })
	}
}
func BenchmarkLongLine(b *testing.B) {
	Opts.FirstMatchOnly = true
	input := "./configure --with-features=huge " +
		"--enable-multibyte " +
		"--enable-rubyinterp=yes " +
		"--enable-pythoninterp=yes " +
		"--with-python-config-dir=/usr/lib/python2.7/config " +
		"--enable-python3interp=yes " +
		"--with-python3-config-dir=/usr/lib/python3.5/config " +
		"--enable-perlinterp=yes " +
		"--enable-luainterp=yes " +
		"--enable-gui=gtk2 " +
		"--enable-cscope " +
		"--prefix=/usr/local"
	benchmarkMatchMany(b, input, "cohuge", "enagtk2", "cwfh")
}

func BenchmarkVeryLongLine(b *testing.B) {
	Opts.FirstMatchOnly = true
	input := `rg --hidden --heading --line-number --color ansi --colors path:none --colors line:none --colors match:fg:red --colors match:style:nobold --ignore-case -g E8/**/.git -g !/home/sergei/e8/code/qux-E8/**/.svn -g !/home/sergei/e8/code/qux-E8/**/.hg -g !/home/sergei/e8/code/qux-E8/**/CVS -g !/home/sergei/e8/code/qux-E8/**/.DS_Store -g !/home/sergei/e8/code/qux-E8/**/node_modules -g !/home/sergei/e8/code/qux-E8/**/bower_components -g !/home/sergei/e8/code/qux-E8/.vscode --max-filesize 17179869184 --no-ignore-parent --follow --regexp \binclude.*nvme\b -- /home/sergei/e8/code/qux-E8\n: 1524066773:0;rg --hidden --heading --line-number --color ansi --colors path:none --colors line:none --colors match:fg:red --colors match:style:nobold --ignore-case -g E8/**/.git -g E8/**/.svn -g !/home/sergei/e8/code/qux-E8/**/.hg -g !/home/sergei/e8/code/qux-E8/**/CVS -g !/home/sergei/e8/code/qux-E8/**/.DS_Store -g !/home/sergei/e8/code/qux-E8/**/node_modules -g !/home/sergei/e8/code/qux-E8/**/bower_components -g !/home/sergei/e8/code/qux-E8/.vscode --max-filesize 17179869184 --no-ignore-parent --follow --regexp \binclude.*nvme\b -- /home/sergei/e8/code/qux-E8\n: 1524066839:0;rg --hidden --heading --line-number --color ansi --colors path:none --colors line:none --colors match:fg:red --colors match:style:nobold --ignore-case -g "E8/**/.git" -g "E8/**/.svn" -g "E8/**/.hg" -g "!/home/sergei/e8/code/qux-E8/**/CVS" -g "!/home/sergei/e8/code/qux-E8/**/.DS_Store" -g "!/home/sergei/e8/code/qux-E8/**/node_modules" -g "!/home/sergei/e8/code/qux-E8/**/bower_components" -g "!/home/sergei/e8/code/qux-E8/.vscode" --max-filesize 17179869184 --no-ignore-parent --follow --regexp "\binclud.*nvme\b" -- /home/sergei/e8/code/qux-E8\n: 1524066903:0;rg --hidden --heading --line-number --color ansi --colors path:none --colors line:none --colors match:fg:red --colors match:style:nobold --ignore-case --max-filesize 17179869184 --follow --regexp "\binclud.*nvme\b" -- /home/sergei/e8/code/qux-E8\n: 1524066987:0;cd ../../`
	benchmarkMatch(b, input, "e8slashmany")
}
