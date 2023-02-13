package DBUtils

import(
	"math"
	"fmt"
	"errors"
	"gonum.org/v1/gonum/mat"
)

var phi1, phi2, lambda1, lambda2, dist1, dist2 float64
var cosPhi1, cosPhi2, const1, const2 float64
const r = 6371.0
const delta = 1.0e-2
const eps = 1.0e-4
const num_it = 1000

func FindNewCoordinates(p1 float64, p2 float64, l1 float64, l2 float64, d1 float64, d2 float64) (float64, float64, error) {

	phi1, phi2, lambda1, lambda2, dist1, dist2 = p1, p2, l1, l2, d1, d2
	cosPhi1, cosPhi2, const1, const2 = math.Cos(phi1), math.Cos(phi2), math.Pow(math.Sin(dist1 / (2*r)), 2), math.Pow(math.Sin(dist2 / (2*r)), 2)	
	
	x_k := mat.NewDense(2, 1, []float64{(phi1 + phi2)/2, (lambda1 + lambda2)/2})
	var x_k_old mat.Dense 
	var f mat.Dense

	var i int
	for i = 0; i < num_it; i++{
		f = function(*x_k)

		if math.Abs(f.At(0, 0)) < eps && math.Abs(f.At(1, 0)) < eps {
			break
		}

		var gaussian mat.LQ
		gaussian.Factorize(mat.NewDense(2, 2, []float64{
			0.5*math.Sin(x_k.At(0,0) - phi1) - cosPhi1*math.Pow(math.Sin((x_k.At(1,0) - lambda1)/2), 2)*math.Sin(x_k.At(0,0)), 0.5*cosPhi1*math.Cos(x_k.At(0,0))*math.Sin(x_k.At(1,0) - lambda1), 
			0.5*math.Sin(x_k.At(0,0) - phi2) - cosPhi2*math.Pow(math.Sin((x_k.At(1,0) - lambda2)/2), 2)*math.Sin(x_k.At(0,0)), 0.5*cosPhi2*math.Cos(x_k.At(0,0))*math.Sin(x_k.At(1,0) - lambda2) }))

		f_norm := math.Pow(f.At(0,0), 2) + math.Pow(f.At(1,0), 2)

		var s mat.Dense
		f.Scale(-1.0, &f)
		gaussian.SolveTo(&s, true, &f)		

		var temp mat.Dense
		temp.Add(x_k, &s)
		temp = function(temp)

		var sigma = 1.0
		for ; (math.Pow(temp.At(0,0), 2) + math.Pow(temp.At(1,0), 2)) > (f_norm - 2*delta*sigma*f_norm) ; sigma /= 2.0 {
			s.Scale(0.5, &s)
			temp.Add(x_k, &s)

			temp = function(temp)
		}
		
		x_k_old.CloneFrom(x_k)
		x_k.Add(x_k, &s)
		
		if mat.Equal(x_k, &x_k_old) || i == num_it-1 {
			return 0, 0, errors.New(fmt.Errorf("Could not find zero-point.").Error());
		}
	}

	return x_k.At(0, 0), x_k.At(1, 0), nil
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