//go:build !unix

package repo

// lockDir does nothing outside Unix for now. Windows support is planned
// later (D5); until then, two inits into the same existing folder at the
// same moment are not protected there.
func lockDir(string) (unlock func(), err error) {
	return func() {}, nil
}
