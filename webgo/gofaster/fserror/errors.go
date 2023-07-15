package fserror

type ErrorFunc func(fsError *FsError)
type FsError struct {
	err     error
	errFunc ErrorFunc
}

func Default() *FsError {
	return &FsError{}
}

func (f *FsError) Error() string {
	return f.err.Error()
}
func (f *FsError) Put(err error) {
	f.check(err)
}

func (f *FsError) check(err error) {
	if err != nil {
		f.err = err
		panic(f)
	}
}

func (f *FsError) Result(errorFunc ErrorFunc) {
	f.errFunc = errorFunc
}

func (f *FsError) ExecResult() {
	f.errFunc(f)

}
