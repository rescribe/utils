package main

import (
	"fmt"
	"testing"
)

func TestSamplePrefixes(t *testing.T) {
	prefixes := Prefixes{
		"1471-Orthographia": {
			"1471-Orthographia-Tortellius_00001.txt",
			"1471-Orthographia-Tortellius_00002.txt",
			"1471-Orthographia-Tortellius_00003.txt",
			"1471-Orthographia-Tortellius_00004.txt",
			"1471-Orthographia-Tortellius_00005.txt",
			"1471-Orthographia-Tortellius_00006.txt",
			"1471-Orthographia-Tortellius_00007.txt",
			"1471-Orthographia-Tortellius_00008.txt",
			"1471-Orthographia-Tortellius_00009.txt",
			"1471-Orthographia-Tortellius_000010.txt",
			"1471-Orthographia-Tortellius_000011.txt",
			"1471-Orthographia-Tortellius_000012.txt",
			"1471-Orthographia-Tortellius_000013.txt",
			"1471-Orthographia-Tortellius_000014.txt",
			"1471-Orthographia-Tortellius_000015.txt",
			"1471-Orthographia-Tortellius_000016.txt",
			"1471-Orthographia-Tortellius_000017.txt",
			"1471-Orthographia-Tortellius_000018.txt",
			"1471-Orthographia-Tortellius_000019.txt",
			"1471-Orthographia-Tortellius_000020.txt",
		},
		"Kallimachos_1509": {
			"Kallimachos_1509-ShipOfFools-Barclay_00121.txt",
			"Kallimachos_1509-ShipOfFools-Barclay_00122.txt",
			"Kallimachos_1509-ShipOfFools-Barclay_00123.txt",
			"Kallimachos_1509-ShipOfFools-Barclay_00124.txt",
			"Kallimachos_1509-ShipOfFools-Barclay_00125.txt",
			"Kallimachos_1509-ShipOfFools-Barclay_00126.txt",
		},
		"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4": {
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_10_59125.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_11_27.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_12_49.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_13_9033333333333333.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_1_415.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_14_6628571428571429.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_16_865.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_17_62.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_18_6366666666666666.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_19_7857142857142857.txt",
			"buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_19_7857142857142857.txt",
		},
	}

	cases := []struct {
		perc     int
		expected []string
	}{
		//{1, []string{""}}, // TODO: fix this; currently causes hang
		{10, []string{"1471-Orthographia-Tortellius_000019.txt", "Kallimachos_1509-ShipOfFools-Barclay_00122.txt", "buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_13_9033333333333333.txt"}},
		{20, []string{"1471-Orthographia-Tortellius_00002.txt", "Kallimachos_1509-ShipOfFools-Barclay_00126.txt", "buckets_1678_DuHAMEL_PhilosophiaVetusEtNova_Vol4_0008_bin0.4-copy_line_1_11_27.txt"}},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%d%%", c.perc), func(t *testing.T) {
			actual := samplePrefixes(c.perc, prefixes)
			if len(c.expected) != len(actual) {
				t.Fatalf("Number of files picked (%d) differs from expected (%d):\nExpected: %s\nActual: %s\n", len(actual), len(c.expected), c.expected, actual)
				return
			}
			for i, v := range c.expected {
				if actual[i] != v {
					t.Fatalf("Difference in expected and actual files (at least in number %d of actual):\n\nExpected:\n%s\n\nActual:\n%s\n", i, c.expected, actual)
				}
			}
		})
	}
}
