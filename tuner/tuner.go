package tuner

import (
	"bufio"
	"bullet/engine"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const (
	NumPSQTWeights = 6 * 64

	RandomDeltaBound float64 = 25

	BasePawnCPValue float64   = 100
	BaseKnightCPValue float64 = 300
	BaseBishopCPValue float64 = 300
	BaseRookCPValue float64   = 500
	BaseQueenCPValue float64  = 850
	BaseKingCPValue float64   = 0

	// Scaling factor so that the conversion from centi-pawns to a
	// probability is more reasonable. E.g., we don't want 50 cp =>
	// 0.99 probability, which is would using a unchanged sigmoid
	// function.
	K              float64 = 0.01
	Epsilon        float64 = 0.00000001
)

var BasePieceValues = [6]float64{
	BasePawnCPValue,
	BaseBishopCPValue,
	BaseKnightCPValue,
	BaseRookCPValue,
	BaseQueenCPValue,
	BaseKingCPValue,
}

func genRandIntWithinSymmetricInterval(bound float64) float64 {
	return (rand.Float64() * 2*bound) - bound
}

func sigmoid(x float64) float64 {
	return 1 / (1 + math.Exp(-x))
}

func evaluatePosition(weights *Weights, pos *Datapoint) (score float64) {
	for _, piece := range pos.Pieces {
		score += float64(piece.Sign) * weights.weights[piece.WeightIdx]
	}
	return score
}

func convertFloatSiceToInt(slice []float64) (ints []int16) {
	for _, float := range slice {
		ints = append(ints, int16(float))
	}
	return ints
}

func prettyPrintPSQT(name string, psqt []int16) {
	fmt.Print("{\n")
	fmt.Print("    // ", name, "\n    ")
	for sq := 0; sq < 64; sq++ {
		if sq > 0 && sq%8 == 0 {
			fmt.Print("\n    ")
		}
		fmt.Printf("%3d, ", psqt[sq])
	}
	fmt.Print("\n},\n")
}

type Piece struct {
	WeightIdx uint16
	Sign      int8
}

type Datapoint struct {
	Outcome  float64 
	Pieces   []Piece
}

func NewDatapoint(pos *engine.Position, outcome float64) Datapoint {
	datapoint := Datapoint{}
	datapoint.Outcome = outcome

	piecesBB := pos.Colors[engine.White] | pos.Colors[engine.Black]

	for piecesBB != 0 {
		sq := engine.GetLSBpos(piecesBB)
		pieceType := pos.GetPieceTypeOnSq(sq)
		pieceColor := pos.GetPieceColorOnSq(sq)
		

		weightIdx := uint16(pieceType)*64 + uint16(engine.FlipSq[pieceColor][sq])
		sign := int8(1)
		if pieceColor == engine.Black {
			sign = -1
		}

		datapoint.Pieces = append(datapoint.Pieces, Piece{WeightIdx: weightIdx, Sign: sign})
		piecesBB &= (piecesBB - 1)
	}

	return datapoint
}

func loadDatapoints(fenFilePath string) (datapoints []Datapoint) {
	dataFile, err := os.OpenFile(fenFilePath, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	defer dataFile.Close()

	scanner := bufio.NewScanner(dataFile)
	datapoints = []Datapoint{}
	pos := engine.Position{}

	// Skip the CSV header
	scanner.Scan()

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		fields := strings.Split(line, ",")

		if len(fields) != 2 {
			panic("malformed datapoint in data file")
		}

		fenField := strings.TrimSpace(fields[0])
		outcomeField := strings.TrimSpace(fields[1])

		pos.LoadFEN(fenField)
		outcome, err := strconv.ParseFloat(outcomeField, 64)

		if err != nil {
			panic(err)
		}

		datapoints = append(datapoints, NewDatapoint(&pos, outcome))
	}

	return datapoints
}

type Weights struct {
	weights               [NumPSQTWeights]float64
	sumOfGradientsSquared [NumPSQTWeights]float64
}

func (weights *Weights) Randomize() {
	for pieceType := engine.Pawn; pieceType < engine.NoType; pieceType++ {
		baseValue := BasePieceValues[pieceType]
		startIdx := pieceType*64

		for sq := 0; sq < 64; sq++ {
			value := baseValue + genRandIntWithinSymmetricInterval(RandomDeltaBound)
			weights.weights[startIdx+sq] = value
		}
	}
}

func (weights *Weights) LoadBaseWeights() {
	for pieceType := engine.Pawn; pieceType < engine.NoType; pieceType++ {
		baseValue := BasePieceValues[pieceType]
		startIdx := pieceType*64
		for sq := 0; sq < 64; sq++ {
			weights.weights[startIdx+sq] = baseValue
		}
	}
}

func (weights *Weights) LoadWeights(PSQT [6][64]int16) {
	for pieceType := engine.Pawn; pieceType < engine.NoType; pieceType++ {
		startIdx := pieceType*64
		for sq := 0; sq < 64; sq++ {
			weights.weights[startIdx+sq] = float64(PSQT[pieceType][sq])
		}
	}
}

func (weights *Weights) CopyWeights(PSQT *[6][64]int16) {
	for pieceType := engine.Pawn; pieceType < engine.NoType; pieceType++ {
		startIdx := pieceType*64
		for sq := 0; sq < 64; sq++ {
			PSQT[pieceType][sq] = int16(weights.weights[startIdx+sq])
		}
	}
}

func (weights *Weights) ComputeMSE(w *Weights, d []Datapoint) (sum float64) {
	for i := 0; i < len(d); i++ {
		datapoint := &d[i]
		y_hat := sigmoid(K*evaluatePosition(w, datapoint))
		diff := datapoint.Outcome - y_hat
		sum += diff * diff
	}
	return sum / float64(len(d))
}

func (weights *Weights) computeAndApplyGradient(learningRate float64, datapoints []Datapoint) {
	gradients := make([]float64, len(weights.weights))

	for i := 0; i < len(datapoints); i++ {
		datapoint := &datapoints[i]
		y_hat := sigmoid(K*evaluatePosition(weights, datapoint))
		term := (datapoint.Outcome - y_hat) * y_hat * (1 - y_hat)

		for _, piece := range datapoint.Pieces {
			gradients[piece.WeightIdx] += term * float64(piece.Sign)
		}
	}

	N := float64(len(datapoints))
	leadingCoeff := (-2 * K) / N

	for i := 0; i < len(gradients); i++ {
		finalGradient := leadingCoeff * gradients[i]
		weights.sumOfGradientsSquared[i] += finalGradient * finalGradient
		sqrtTerm := math.Sqrt(weights.sumOfGradientsSquared[i]+Epsilon)
		weights.weights[i] -= learningRate * finalGradient / sqrtTerm
	}
}

func (weights *Weights) DisplayWeights() {
	prettyPrintPSQT("Pawn PST", convertFloatSiceToInt(weights.weights[0:64]))
	prettyPrintPSQT("Knight PST", convertFloatSiceToInt(weights.weights[64:128]))
	prettyPrintPSQT("Bishop PST", convertFloatSiceToInt(weights.weights[128:192]))
	prettyPrintPSQT("Rook PST", convertFloatSiceToInt(weights.weights[192:256]))
	prettyPrintPSQT("Queen PST", convertFloatSiceToInt(weights.weights[256:320]))
	prettyPrintPSQT("King PST", convertFloatSiceToInt(weights.weights[320:384]))
}

func (weights *Weights) TuneWeights(dataFile string, learningRate float64, iterations, recordErrEveryNth int) {
	datapoints := loadDatapoints(dataFile)
	weights.sumOfGradientsSquared = [NumPSQTWeights]float64{}

	beforeErr := weights.ComputeMSE(weights, datapoints)
	errors := []float64{beforeErr}

	for i := 0; i < iterations; i++ {
		weights.computeAndApplyGradient(learningRate, datapoints)
		fmt.Printf("Completed iteration %d/%d\n", i+1, iterations)
		
		if i > 0 && i % recordErrEveryNth == 0 {
			errors = append(errors, weights.ComputeMSE(weights, datapoints))
		}
	}

	errors = append(errors, weights.ComputeMSE(weights, datapoints))
	file, err := os.Create("errors.txt")
	if err != nil {
		fmt.Println("Couldn't create \"errors.txt\" to store recored error rates")
	} else {
		fmt.Println("Storing error rates in errors.txt")
	}

	for _, err := range errors {
		_, e := file.WriteString(fmt.Sprintf("%f\n", err))
		if e != nil {
			panic(e)
		}
	}

	fmt.Println("Before MSE:", beforeErr)
	fmt.Println("After MSE:", weights.ComputeMSE(weights, datapoints))

	weights.DisplayWeights()
}
