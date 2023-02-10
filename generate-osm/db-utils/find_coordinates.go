package DBUtils

import(
	"math"
	"fmt"
	"gonum.org/v1/gonum/mat"
)

var phi1, phi2, lambda1, lambda2, dist1, dist2 float64
var cosPhi1, cosPhi2, const1, const2 float64
const r = 6371.0
const delta = 1.0e-2

func FindNewCoordinates(p1 float64, p2 float64, l1 float64, l2 float64, d1 float64, d2 float64) (float64, float64) {

	phi1, phi2, lambda1, lambda2, dist1, dist2 = p1, p2, l1, l2, d1, d2
	cosPhi1, cosPhi2, const1, const2 = math.Cos(phi1), math.Cos(phi2), math.Pow(math.Sin(dist1 / (2*r)), 2), math.Pow(math.Sin(dist2 / (2*r)), 2)	
	
	result := mat.NewDense(2, 1, []float64{phi1, lambda2})
	var f mat.Dense

	for ; true; {
		f = function(*result)

		fmt.Printf("%f, %f \n", f.At(0,0), f.At(1,0))

		if math.Abs(f.At(0, 0)) < 1.0e-3 && math.Abs(f.At(1, 0)) < 1.0e-3 {
			break
		}

		var gaussian mat.LQ
		gaussian.Factorize(mat.NewDense(2, 2, []float64{
			0.5 * math.Sin(result.At(0, 0) - phi1) - cosPhi1*math.Pow(math.Sin((result.At(1, 0) - lambda1)/2), 2)*math.Sin(result.At(0, 0)), 0.5 * cosPhi1*math.Cos(result.At(0, 0))*math.Sin(result.At(1, 0) - lambda1), 
			0.5 * math.Sin(result.At(0, 0) - phi2) - cosPhi2*math.Pow(math.Sin((result.At(1, 0) - lambda2)/2), 2)*math.Sin(result.At(0, 0)), 0.5 * cosPhi2*math.Cos(result.At(0, 0))*math.Sin(result.At(1, 0) - lambda2) }))

		f_norm := math.Pow(f.At(0,0), 2) + math.Pow(f.At(1,0), 2)

		var s mat.Dense
		f.Scale(-1.0, &f)
		gaussian.SolveTo(&s, false, &f)		

		var temp mat.Dense
		temp.Add(result, &s)

		var sigma = 1.0
		for ; (math.Pow(temp.At(0,0), 2) + math.Pow(temp.At(1,0), 2)) > (f_norm - 2*delta*sigma*f_norm) ; sigma /= 2.0 {
			s.Scale(0.5, &s)
			temp.Add(result, &s)

			temp = function(temp)

			if sigma/2 == 0.0 {
				break
			}
		}
		
		result.Add(result, &s)
	}

	return result.At(0, 0), result.At(1, 0)
}

func function (in mat.Dense) mat.Dense {
	if rows, cols := in.Dims(); rows != 2 && cols != 1 {
		return *mat.NewDense(2, 1, nil)
	}

	const1 := math.Pow(math.Sin(dist1 / (2*r)), 2)
	const2 := math.Pow(math.Sin(dist2 / (2*r)), 2)	

	return *mat.NewDense(2, 1, []float64{
		math.Pow(math.Sin((in.At(0, 0) - phi1)/2), 2) + math.Cos(in.At(0, 0))*cosPhi1*math.Pow(math.Sin((in.At(1, 0) - lambda1)/2), 2) - const1,
		math.Pow(math.Sin((in.At(0, 0) - phi2)/2), 2) + math.Cos(in.At(0, 0))*cosPhi2*math.Pow(math.Sin((in.At(1, 0) - lambda2)/2), 2) - const2 })
}