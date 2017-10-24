
type PetStore struct {
    DryRun        bool `commander:"long=dru-run"`
    StoreLocation string `commander:"long=store-location"`
}

func (application Application) Create(
