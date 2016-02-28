package lcs

import (
	"log"
	"strings"
)

// <summary>
// This Class implements the Difference Algorithm published in
// "An O(ND) Difference Algorithm and its Variations" by Eugene Myers
// Algorithmica Vol. 1 No. 2, 1986, p 251.
//
// There are many C, Java, Lisp implementations but available they all seem to come
// from the same source (diffutils) that is under the (unfree) GNU public License
// and cannot be reused as a sourcecode for a commercial application.
// There are very old C implementations that use other (worse) algorithms.
// Microsoft also published sourcecode of a diff-tool (windiff) that uses some tree data.
// Also, a direct transfer from a C source to C# is not easy because there is a lot of pointer
// arithmetic in the typical C solutions and i need a managed solution.
// These are the reasons why I implemented the original published algorithm from the scratch and
// make it avaliable without the GNU license limitations.
// I do not need a high performance diff tool because it is used only sometimes.
// I will do some performace tweaking when needed.
//
// The algorithm itself is comparing 2 arrays of numbers so when comparing 2 text documents
// each line is converted into a (hash) number. See DiffText().
//
// Some chages to the original algorithm:
// The original algorithm was described using a recursive approach and comparing zero indexed arrays.
// Extracting sub-arrays and rejoining them is very performance and memory intensive so the same
// (readonly) data arrays are passed arround together with their lower and upper bounds.
// This circumstance makes the LCS and SMS functions more complicate.
// I added some code to the LCS function to get a fast response on sub-arrays that are identical,
// completely deleted or inserted.
//
// The result from a comparisation is stored in 2 arrays that flag for modified (deleted or inserted)
// lines in the 2 data arrays. These bits are then analysed to produce a array of Operation objects.
//
// Further possible optimizations:
// (first rule: don't do it second: don't do it yet)
// The arrays DataA and DataB are passed as parameters, but are never changed after the creation
// so they can be members of the class to a the paramter overhead.
// In SMS is a lot of boundary arithmetic in the for-D and for-k loops that can be done by increment
// and decrement of local variables.
// The DownVector and UpVector arrays are alywas created and destroyed each time the SMS gets called.
// It is possible to reuse tehm when transfering them to members of the class.
// See TODO: hints.
//
// diff.cs: A port of the algorythm to C#
// Copyright (c) by Matthias Hertel, http://www.mathertel.de
// This work is licensed under a BSD style license. See http://www.mathertel.de/License.aspx
//
// Changes:
// 2002.09.20 There was a "hang" in some situations.
// Now I undestand a little bit more of the SMS algorithm.
// There have been overlapping boxes that where analyzed partial differently.
// One return-point is enough.
// A assertion was added in CreateDiffs when in debug-mode, that counts the number of equal (no modified) lines in both arrays.
// They must be identical.
//
// 2003.02.07 Out of bounds error in the Up/Down vector arrays in some situations.
// The two vetors are now accessed using different offsets that are adjusted using the start k-Line.
// A test case is added.
//
// 2006.03.05 Some documentation and a direct Diff entry point.
//
// 2006.03.08 Refactored the API to static methods on the Diff class to make usage simpler.
// 2006.03.10 using the standard Debug class for self-test now.
//            compile with: csc /target:exe /out:diffTest.exe /d:DEBUG /d:TRACE /d:SELFTEST Diff.cs
// 2007.01.06 license agreement changed to a BSD style license.
// 2007.06.03 added the Optimize method.
// 2007.09.23 UpVector and DownVector optimization by Jan Stoklasa ().
// 2008.05.31 Adjusted the testing code that failed because of the Optimize method (not a bug in the diff algorithm).
// 2008.10.08 Fixing a test case and adding a new test case.
// </summary>

// <summary>details of one difference.</summary>
type Operation struct {
	// <summary>Start Line number in Data A.</summary>
	StartA int
	// <summary>Start Line number in Data B.</summary>
	StartB int

	// <summary>Number of changes in Data A.</summary>
	DeleteA int
	// <summary>Number of changes in Data B.</summary>
	InsertB int
} // Operation

// <summary>
// Shortest Middle Snake Return Data
// </summary>
type SMSRD struct {
	x, y int
}

// <summary>
// Find the difference in 2 texts, comparing by textlines.
// </summary>
// <param name="TextA">A-version of the text (usualy the old one)</param>
// <param name="TextB">B-version of the text (usualy the new one)</param>
// <returns>Returns a array of Items that describe the differences.</returns>
func DiffText(TextA string, TextB string) []Operation {
	return (DiffTextOpts(TextA, TextB, false, false, false))
}

type Hashtable map[string]int

// <summary>
// Find the difference in 2 text documents, comparing by textlines.
// The algorithm itself is comparing 2 arrays of numbers so when comparing 2 text documents
// each line is converted into a (hash) number. This hash-value is computed by storing all
// textlines into a common hashtable so i can find dublicates in there, and generating a
// new number each time a new textline is inserted.
// </summary>
// <param name="TextA">A-version of the text (usualy the old one)</param>
// <param name="TextB">B-version of the text (usualy the new one)</param>
// <param name="trimSpace">When set to true, all leading and trailing whitespace characters are stripped out before the comparation is done.</param>
// <param name="ignoreSpace">When set to true, all whitespace characters are converted to a single space character before the comparation is done.</param>
// <param name="ignoreCase">When set to true, all characters are converted to their lowercase equivivalence before the comparation is done.</param>
// <returns>Returns a array of Items that describe the differences.</returns>
func DiffTextOpts(TextA string, TextB string, trimSpace bool, ignoreSpace bool, ignoreCase bool) []Operation {
	// prepare the input-text and convert to comparable numbers.
	// Hashtable := new Hashtable(len(TextA) + len(TextB))
	h := Hashtable{}

	// The A-Version of the data (original data) to be compared.
	DataA := NewDiffData(DiffCodes(TextA, h, trimSpace, ignoreSpace, ignoreCase))

	// The B-Version of the data (modified data) to be compared.
	DataB := NewDiffData(DiffCodes(TextB, h, trimSpace, ignoreSpace, ignoreCase))

	h = nil // free up hashtable memory (maybe)

	MAX := DataA.Length + DataB.Length + 1
	// vector for the (0,0) to (x,y) search
	DownVector := make([]int, 2*MAX+2)
	// vector for the (u,v) to (N,M) search
	UpVector := make([]int, 2*MAX+2)

	LCS(DataA, 0, DataA.Length, DataB, 0, DataB.Length, DownVector, UpVector)

	Optimize(DataA)
	Optimize(DataB)
	return CreateDiffs(DataA, DataB)
} // DiffText

// <summary>
// If a sequence of modified lines starts with a line that contains the same content
// as the line that appends the changes, the difference sequence is modified so that the
// appended line and not the starting line is marked as modified.
// This leads to more readable diff sequences when comparing text files.
// </summary>
// <param name="Data">A Diff data buffer containing the identified changes.</param>
func Optimize(Data DiffData) {
	var StartPos, EndPos int

	StartPos = 0
	for StartPos < Data.Length {
		for StartPos < Data.Length && (Data.modified[StartPos] == false) {
			StartPos++
		}
		EndPos = StartPos
		for EndPos < Data.Length && (Data.modified[EndPos] == true) {
			EndPos++
		}
		if (EndPos < Data.Length) && (Data.data[StartPos] == Data.data[EndPos]) {
			Data.modified[StartPos] = false
			Data.modified[EndPos] = true
		} else {
			StartPos = EndPos
		} // if
	} // while
} // Optimize

// <summary>
// Find the difference in 2 arrays of integers.
// </summary>
// <param name="ArrayA">A-version of the numbers (usualy the old one)</param>
// <param name="ArrayB">B-version of the numbers (usualy the new one)</param>
// <returns>Returns a array of Items that describe the differences.</returns>
func DiffInt(ArrayA []int, ArrayB []int) []Operation {
	// The A-Version of the data (original data) to be compared.
	DataA := NewDiffData(ArrayA)

	// The B-Version of the data (modified data) to be compared.
	DataB := NewDiffData(ArrayB)

	MAX := DataA.Length + DataB.Length + 1
	// vector for the (0,0) to (x,y) search
	DownVector := make([]int, 2*MAX+2)
	// vector for the (u,v) to (N,M) search
	UpVector := make([]int, 2*MAX+2)

	LCS(DataA, 0, DataA.Length, DataB, 0, DataB.Length, DownVector, UpVector)
	return CreateDiffs(DataA, DataB)
} // Diff

// <summary>
// This function converts all textlines of the text into unique numbers for every unique textline
// so further work can work only with simple numbers.
// </summary>
// <param name="aText">the input text</param>
// <param name="h">This extern initialized hashtable is used for storing all ever used textlines.</param>
// <param name="trimSpace">ignore leading and trailing space characters</param>
// <returns>a array of integers.</returns>
func DiffCodes(str string, h Hashtable, trimSpace bool, ignoreSpace bool, ignoreCase bool) []int {
	lastUsedCode := len(h)
	Codes := make([]int, len(str))
	for i, s := range strings.Split(str, "") {
		if aCode, ok := h[s]; ok {
			Codes[i] = aCode
		} else {
			lastUsedCode++
			h[s] = lastUsedCode
			Codes[i] = lastUsedCode
		}
	}
	log.Printf("%s -> %v", str, Codes)
	return (Codes)
}

// <summary>
// This is the algorithm to find the Shortest Middle Snake (SMS).
// </summary>
// <param name="DataA">sequence A</param>
// <param name="LowerA">lower bound of the actual range in DataA</param>
// <param name="UpperA">upper bound of the actual range in DataA (exclusive)</param>
// <param name="DataB">sequence B</param>
// <param name="LowerB">lower bound of the actual range in DataB</param>
// <param name="UpperB">upper bound of the actual range in DataB (exclusive)</param>
// <param name="DownVector">a vector for the (0,0) to (x,y) search. Passed as a parameter for speed reasons.</param>
// <param name="UpVector">a vector for the (u,v) to (N,M) search. Passed as a parameter for speed reasons.</param>
// <returns>a MiddleSnakeData record containing x,y and u,v</returns>
func SMS(DataA DiffData, LowerA, UpperA int, DataB DiffData, LowerB, UpperB int, DownVector []int, UpVector []int) SMSRD {

	MAX := DataA.Length + DataB.Length + 1

	DownK := LowerA - LowerB // the k-line to start the forward search
	UpK := UpperA - UpperB   // the k-line to start the reverse search

	Delta := (UpperA - LowerA) - (UpperB - LowerB)
	oddDelta := (Delta & 1) != 0

	// The vectors in the publication accepts negative indexes. the vectors implemented here are 0-based
	// and are access using a specific offset: UpOffset UpVector and DownOffset for DownVektor
	DownOffset := MAX - DownK
	UpOffset := MAX - UpK

	MaxD := ((UpperA - LowerA + UpperB - LowerB) / 2) + 1

	// Debug.Write(2, "SMS", String.Format("Search the box: A[{0}-{1}] to B[{2}-{3}]", LowerA, UpperA, LowerB, UpperB))

	// init vectors
	DownVector[DownOffset+DownK+1] = LowerA
	UpVector[UpOffset+UpK-1] = UpperA

	for D := 0; D <= MaxD; D++ {

		// Extend the forward path.
		for k := DownK - D; k <= DownK+D; k += 2 {
			// Debug.Write(0, "SMS", "extend forward path " + k.ToString())

			// find the only or better starting point
			var x, y int
			if k == DownK-D {
				x = DownVector[DownOffset+k+1] // down
			} else {
				x = DownVector[DownOffset+k-1] + 1 // a step to the right
				if (k < DownK+D) && (DownVector[DownOffset+k+1] >= x) {
					x = DownVector[DownOffset+k+1] // down
				}
			}
			y = x - k

			// find the end of the furthest reaching forward D-path in diagonal k.
			for x < UpperA && (y < UpperB) && (DataA.data[x] == DataB.data[y]) {
				x++
				y++
			}
			DownVector[DownOffset+k] = x

			// overlap ?
			if oddDelta && (UpK-D < k) && (k < UpK+D) {
				if UpVector[UpOffset+k] <= DownVector[DownOffset+k] {
					return SMSRD{
						x: DownVector[DownOffset+k],
						y: DownVector[DownOffset+k] - k,
					}
				} // if
			} // if

		} // for k

		// Extend the reverse path.
		for k := UpK - D; k <= UpK+D; k += 2 {
			// Debug.Write(0, "SMS", "extend reverse path " + k.ToString())

			// find the only or better starting point
			var x, y int
			if k == UpK+D {
				x = UpVector[UpOffset+k-1] // up
			} else {
				x = UpVector[UpOffset+k+1] - 1 // left
				if (k > UpK-D) && (UpVector[UpOffset+k-1] < x) {
					x = UpVector[UpOffset+k-1] // up
				}
			} // if
			y = x - k

			for x > LowerA && (y > LowerB) && (DataA.data[x-1] == DataB.data[y-1]) {
				x--
				y-- // diagonal
			}
			UpVector[UpOffset+k] = x

			// overlap ?
			if !oddDelta && (DownK-D <= k) && (k <= DownK+D) {
				if UpVector[UpOffset+k] <= DownVector[DownOffset+k] {
					// ret.u = UpVector[UpOffset + k]     // 2002.09.20: no need for 2 points
					// ret.v = UpVector[UpOffset + k] - k
					return SMSRD{
						x: DownVector[DownOffset+k],
						y: DownVector[DownOffset+k] - k,
					}
				} // if
			} // if

		} // for k

	} // for D

	panic("the algorithm should never come here.")
} // SMS

// <summary>
// This is the divide-and-conquer implementation of the longes common-subsequence (LCS)
// algorithm.
// The published algorithm passes recursively parts of the A and B sequences.
// To a copying these arrays the lower and upper bounds are passed while the sequences stay constant.
// </summary>
// <param name="DataA">sequence A</param>
// <param name="LowerA">lower bound of the actual range in DataA</param>
// <param name="UpperA">upper bound of the actual range in DataA (exclusive)</param>
// <param name="DataB">sequence B</param>
// <param name="LowerB">lower bound of the actual range in DataB</param>
// <param name="UpperB">upper bound of the actual range in DataB (exclusive)</param>
// <param name="DownVector">a vector for the (0,0) to (x,y) search. Passed as a parameter for speed reasons.</param>
// <param name="UpVector">a vector for the (u,v) to (N,M) search. Passed as a parameter for speed reasons.</param>
func LCS(DataA DiffData, LowerA, UpperA int, DataB DiffData, LowerB, UpperB int, DownVector []int, UpVector []int) {
	// Debug.Write(2, "LCS", String.Format("Analyse the box: A[{0}-{1}] to B[{2}-{3}]", LowerA, UpperA, LowerB, UpperB))

	// Fast walkthrough equal lines at the start
	for LowerA < UpperA && LowerB < UpperB && DataA.data[LowerA] == DataB.data[LowerB] {
		LowerA++
		LowerB++
	}

	// Fast walkthrough equal lines at the end
	for LowerA < UpperA && LowerB < UpperB && DataA.data[UpperA-1] == DataB.data[UpperB-1] {
		UpperA--
		UpperB--
	}

	if LowerA == UpperA {
		// mark as inserted lines.
		for LowerB < UpperB {
			DataB.modified[LowerB] = true
			LowerB++
		}
	} else if LowerB == UpperB {
		// mark as deleted lines.
		for LowerA < UpperA {
			DataA.modified[LowerA] = true
			LowerA++

		}

	} else {
		// Find the middle snakea and length of an optimal path for A and B
		smsrd := SMS(DataA, LowerA, UpperA, DataB, LowerB, UpperB, DownVector, UpVector)
		// Debug.Write(2, "MiddleSnakeData", String.Format("{0},{1}", smsrd.x, smsrd.y))

		// The path is from LowerX to (x,y) and (x,y) to UpperX
		LCS(DataA, LowerA, smsrd.x, DataB, LowerB, smsrd.y, DownVector, UpVector)
		LCS(DataA, smsrd.x, UpperA, DataB, smsrd.y, UpperB, DownVector, UpVector) // 2002.09.20: no need for 2 points
	}
} // LCS()

type ArrayList []int

// <summary>Scan the tables of which lines are inserted and deleted,
// producing an edit script in forward order.
// </summary>
// dynamic array
func CreateDiffs(DataA DiffData, DataB DiffData) []Operation {
	var op Operation
	var result []Operation

	var StartA, StartB int
	var LineA, LineB int

	LineA = 0
	LineB = 0
	for LineA < DataA.Length || LineB < DataB.Length {
		if (LineA < DataA.Length) && (!DataA.modified[LineA]) && (LineB < DataB.Length) && (!DataB.modified[LineB]) {
			// equal lines
			LineA++
			LineB++

		} else {
			// maybe deleted and/or inserted lines
			StartA = LineA
			StartB = LineB

			for LineA < DataA.Length && (LineB >= DataB.Length || DataA.modified[LineA]) {
				// for LineA < DataA.Length && DataA.modified[LineA]
				LineA++
			}
			for LineB < DataB.Length && (LineA >= DataA.Length || DataB.modified[LineB]) {
				// for LineB < DataB.Length && DataB.modified[LineB]
				LineB++
			}
			if (StartA < LineA) || (StartB < LineB) {
				// store a new difference-Operation
				op = Operation{}
				op.StartA = StartA
				op.StartB = StartB
				op.DeleteA = LineA - StartA
				op.InsertB = LineB - StartB
				result = append(result, op)
			} // if
		} // if
	} // while

	return result
}

// <summary>Data on one input file being compared.
// </summary>

func NewDiffData(initData []int) DiffData {
	return DiffData{
		Length:   len(initData),
		data:     initData,
		modified: make([]bool, len(initData)+2),
	}
}

type DiffData struct {
	// Number of elements (lines).
	Length int
	// Buffer of numbers that will be compared.
	data []int
	// Array of booleans that flag for modified data.
	// This is the result of the diff.
	// This means DeleteA in the first Data or inserted in the second Data.
	modified []bool
}
