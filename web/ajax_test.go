package web

/*
type (
	triple struct {
		p    float64
		i, n int
	}
)

func TestDetermineTest(t *testing.T) {
	tests := []triple{{0.9, 0, 100}, {0.001, 12, 100}, {0, 99, 100}, {0.5, 50, 100}, {0.3, 70, 100}}
	for _, c := range tests {
		if e := testDetermineTest(c.p, c.i, c.n); e != nil {
			t.Error(e)
		}
	}

}

func testDetermineTest(p float64, i, n int) error {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	s := project.NewSubmission(bson.NewObjectId(), "user", project.FILE_MODE, 10000)
	if e := db.Add(db.SUBMISSIONS, s); e != nil {
		return e
	}
	cf := &project.File{
		Id:    bson.NewObjectId(),
		SubId: s.Id,
		Name:  "Testee",
		Time:  s.Time + 25,
	}
	r := &jacoco.Result{
		Id:     bson.NewObjectId(),
		Type:   jacoco.NAME,
		FileId: cf.Id,
	}
	fs := make([]*project.File, n)
	for j := 0; j < n; j++ {
		fs[j] = &project.File{
			Id:    bson.NewObjectId(),
			SubId: s.Id,
			Name:  "Test",
			Time:  s.Time + int64((i+12)*23),
		}
		if e := db.Add(db.FILES, fs[j]); e != nil {
			return e
		}
	}
	cf.Results = bson.M{r.Type + "-" + fs[i].Id.Hex(): r.Id}
	if e := db.Add(db.FILES, cf); e != nil {
		return e
	}
	r.TestId = fs[i].Id
	if e := db.Add(db.RESULTS, r); e != nil {
		return e
	}
	f, e := determineTest(fs, r.Type, p)
	if e != nil {
		return e
	}
	if f.Id != fs[i].Id {
		return fmt.Errorf("incorrect file %s expected %s", f.Id, fs[i].Id)
	}
	return nil
}

func TestDetermineSrc(t *testing.T) {
	tests := []triple{{0.9, 0, 100}, {0.001, 12, 100}, {0, 99, 100}, {0.5, 50, 100}, {0.3, 70, 100}}
	for _, c := range tests {
		if e := testDetermineSrc(c.p, c.i, c.n); e != nil {
			t.Error(e)
		}
	}

}

func testDetermineSrc(p float64, i, n int) error {
	db.Setup(db.TEST_CONN)
	defer db.DeleteDB(db.TEST_DB)
	s := project.NewSubmission(bson.NewObjectId(), "user", project.FILE_MODE, 10000)
	if e := db.Add(db.SUBMISSIONS, s); e != nil {
		return e
	}
	cf := &project.File{
		Id:    bson.NewObjectId(),
		SubId: s.Id,
		Name:  "Test",
		Time:  s.Time + 100,
	}
	r := &jacoco.Result{
		Id:     bson.NewObjectId(),
		Type:   jacoco.NAME,
		TestId: cf.Id,
	}
	fs := make([]*project.File, n)
	for j := 0; j < n; j++ {
		fs[j] = &project.File{
			Id:    bson.NewObjectId(),
			SubId: s.Id,
			Name:  "Testee",
			Time:  s.Time + int64((i+12)*23),
		}
		if i == j {
			fs[j].Results = bson.M{r.Type + "-" + cf.Id.Hex(): r.Id}
		}
		if e := db.Add(db.FILES, fs[j]); e != nil {
			return e
		}
	}
	r.FileId = fs[i].Id
	if e := db.Add(db.FILES, cf); e != nil {
		return e
	}
	if e := db.Add(db.RESULTS, r); e != nil {
		return e
	}
	f, e := determineSrc(fs, r.Type, cf.Name, p)
	if e != nil {
		return e
	}
	if f.Id != fs[i].Id {
		return fmt.Errorf("incorrect file %s expected %s", f.Id, fs[i].Id)
	}
	return nil
}
*/
