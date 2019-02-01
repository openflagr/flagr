package lists

// Combinations returns all possible unique selections of size `pick` of a list
// of strings for which order does not matter
//
// an example:
//
//     Combinations([cat, dog, bird], 2):
//       [cat] -> Combinations([dog, bird], 1)
//         [cat, dog]
//         [cat, bird]
//       [dog] -> Combinations([bird], 1)
//         [dog, bird]
//       [bird] -> Combinations([], 0)
//         n/a
//
func Combinations(list []string, pick int) (all [][]string) {
	switch pick {
	case 0:
		// nothing to do
	case 1:
		for i := range list {
			all = append(all, list[i:i+1])
		}
	default:
		// we recursively find combinations by taking each item in the list
		// and then finding the combinations at (pick-1) for the remaining
		// items in the list
		// the reason we start at [i+1:], is because the order of the items in
		// the list doesn't matter, so this will remove all the duplicates we
		// would get otherwise
		for i := range list {
			for _, next := range Combinations(list[i+1:], pick-1) {
				all = append(all, append([]string{list[i]}, next...))
			}
		}
	}
	return all
}
