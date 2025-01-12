package engine

const (
	InfinityCPValue int16 = 10_000
	DrawCPValue     int16 = 0
)

var PieceSquareTable = [6][64]int16{
	{
		// Pawn PST
		100, 100, 100, 100, 100, 100, 100, 100,
		156, 141, 145, 140, 136, 134, 119, 128,
		130, 128, 120, 129, 128, 114, 118, 126,
		 95,  95,  85,  91,  84,  87,  85,  81,
		 87,  87,  76,  88,  82,  77,  77,  74,
		 83,  88,  74,  76,  77,  73,  86,  72,
		 79,  86,  70,  65,  59,  81,  83,  66,
		100, 100, 100, 100, 100, 100, 100, 100,
	},
	{
		// Knight PST
		280, 295, 293, 291, 290, 293, 297, 294,
		272, 292, 291, 308, 291, 297, 286, 283,
		287, 302, 322, 329, 309, 307, 296, 282,
		306, 296, 308, 323, 308, 313, 296, 306,
		282, 303, 309, 304, 300, 311, 291, 275,
		279, 286, 291, 297, 297, 296, 293, 276,
		288, 269, 284, 292, 290, 286, 293, 291,
		276, 275, 275, 274, 273, 279, 270, 275,
	},
	{
		// Bishop PST
		299, 306, 299, 299, 297, 301, 297, 299,
		302, 317, 304, 300, 307, 302, 298, 286,
		313, 310, 314, 326, 315, 321, 309, 321,
		298, 312, 314, 326, 324, 313, 309, 299,
		297, 311, 319, 324, 316, 309, 322, 305,
		298, 316, 316, 315, 311, 318, 317, 306,
		307, 308, 311, 306, 313, 306, 316, 297,
		287, 287, 298, 292, 296, 291, 298, 290,
	},
	{
		// Rook PST
		513, 513, 508, 511, 501, 503, 499, 498,
		509, 514, 516, 520, 512, 511, 499, 503,
		508, 511, 515, 516, 508, 504, 497, 493,
		495, 503, 501, 505, 493, 490, 484, 490,
		485, 487, 503, 499, 497, 489, 479, 479,
		476, 489, 486, 490, 485, 482, 491, 481,
		474, 484, 491, 493, 486, 493, 481, 441,
		481, 489, 493, 494, 490, 492, 477, 474,
	},
	{
		// Queen PST
		856, 851, 847, 849, 855, 854, 852, 851,
		835, 837, 860, 857, 853, 861, 848, 852,
		860, 849, 852, 862, 871, 871, 867, 861,
		840, 851, 857, 868, 852, 863, 859, 859,
		850, 847, 856, 854, 861, 862, 859, 857,
		850, 850, 849, 854, 858, 854, 862, 843,
		847, 840, 858, 856, 858, 855, 839, 844,
		846, 847, 845, 854, 848, 839, 843, 848,
	},
	{
		// King PST
		  1,   0,   0,   1,   2,   1,  -1,   0,
		  2,   6,   6,   4,   1,   7,  12,  -1,
		  4,  18,  18,  13,  11,  14,  17,   0,
		  1,  16,  17,  17,  20,  11,  16,  -5,
		 -4,   2,  12,   7,  10,   6,   1,  -5,
		 -4,   5,   4,   2,  -1,  -3,   1, -16,
		-12, -12,  -2, -15, -17, -10,  -1,  -4,
		-23,  -9, -14, -34, -14, -27,  -1, -14,
	},
}

var FlipSq = [2][64]uint8{
	{
		A8, B8, C8, D8, E8, F8, G8, H8,
		A7, B7, C7, D7, E7, F7, G7, H7,
		A6, B6, C6, D6, E6, F6, G6, H6,
		A5, B5, C5, D5, E5, F5, G5, H5,
		A4, B4, C4, D4, E4, F4, G4, H4,
		A3, B3, C3, D3, E3, F3, G3, H3,
		A2, B2, C2, D2, E2, F2, G2, H2,
		A1, B1, C1, D1, E1, F1, G1, H1,
	},
	{
		A1, B1, C1, D1, E1, F1, G1, H1,
		A2, B2, C2, D2, E2, F2, G2, H2,
		A3, B3, C3, D3, E3, F3, G3, H3,
		A4, B4, C4, D4, E4, F4, G4, H4,
		A5, B5, C5, D5, E5, F5, G5, H5,
		A6, B6, C6, D6, E6, F6, G6, H6,
		A7, B7, C7, D7, E7, F7, G7, H7,
		A8, B8, C8, D8, E8, F8, G8, H8,
	},
}

func EvaluatePosition(pos *Position) int16 {
	return pos.Scores[pos.Side] - pos.Scores[pos.Side^1]
}
