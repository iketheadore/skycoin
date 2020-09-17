package skycoin

import (
	"fmt"

	"github.com/blang/semver"

	"github.com/skycoin/skycoin/src/cipher"
	"github.com/skycoin/skycoin/src/util/logging"
	"github.com/skycoin/skycoin/src/visor"
	"github.com/skycoin/skycoin/src/visor/dbutil"
)

type dbAction uint

const (
	doNothing dbAction = iota
	doCheckDB
	doResetCorruptDB
)

// dbCheckConfig contains the parameters for verifying db
type dbCheckConfig struct {
	// ForceVerify force verify DB
	ForceVerify bool
	// ResetCorruptDB reset the DB if it is corrupted
	ResetCorruptDB bool
	// // AppVersion is the current wallet version
	// AppVersion *semver.Version
	// DBCheckpointVersion is the check point db version
	DBCheckpointVersion *semver.Version
}

//go:generate mockery -name dbCheckCorruptResetter -case underscore -inpkg -testonly
type dbCheckCorruptResetter interface {
	CheckDatabase(db *dbutil.DB) error
	ResetCorruptDB(db *dbutil.DB) (*dbutil.DB, error)
	GetDBVersion(db *dbutil.DB) (*semver.Version, error)
	SetDBVersion(db *dbutil.DB, v *semver.Version) error
}

type dbVerify struct {
	blockchainPubkey cipher.PubKey
	logger           *logging.Logger
	quit             chan struct{}
}

func (dv dbVerify) CheckDatabase(db *dbutil.DB) error {
	if err := visor.CheckDatabase(db, dv.blockchainPubkey, dv.quit); err != nil {
		if err != visor.ErrVerifyStopped {
			dv.logger.WithError(err).Error("visor.CheckDatabase failed")
		}
		return err
	}
	return nil
}

func (dv *dbVerify) ResetCorruptDB(db *dbutil.DB) (*dbutil.DB, error) {
	dv.logger.Info("Checking database and resetting if corrupted")
	newDB, err := visor.ResetCorruptDB(db, dv.blockchainPubkey, dv.quit)
	if err != nil {
		if err != visor.ErrVerifyStopped {
			dv.logger.WithError(err).Error("visor.ResetCorruptDB failed")
		}
		return nil, err
	}

	return newDB, nil
}

func (dv *dbVerify) SetDBVersion(db *dbutil.DB, v *semver.Version) error {
	if err := visor.SetDBVersion(db, *v); err != nil {
		if err != visor.ErrVerifyStopped {
			dv.logger.WithError(err).Error("visor.ResetCorruptDB failed")
		}
		return err
	}
	return nil
}

func (dv dbVerify) GetDBVersion(db *dbutil.DB) (*semver.Version, error) {
	dbVersion, err := visor.GetDBVersion(db)
	if err != nil {
		dv.logger.WithError(err).Error("visor.GetDBVersion failed")
		return nil, err
	}

	if dbVersion == nil {
		dv.logger.Info("DB version not found in DB")
	} else {
		dv.logger.Infof("DB version: %s", dbVersion)
	}
	return dbVersion, nil
}

func checkAndUpdateDB(db *dbutil.DB, c dbCheckConfig, dv dbCheckCorruptResetter) (*dbutil.DB, error) {
	dbVersion, err := dv.GetDBVersion(db)
	if err != nil {
		return nil, err
	}

	action, err := checkDBVersion(c, dbVersion)
	if err != nil {
		return nil, err
	}

	switch action {
	case doCheckDB:
		if err := dv.CheckDatabase(db); err != nil {
			return nil, err
		}
	case doResetCorruptDB:
		// Check the database integrity and recreate it if necessary
		newDB, err := dv.ResetCorruptDB(db)
		if err != nil {
			return nil, err
		}
		db = newDB
	case doNothing:
		return db, nil
	}

	// DB version won't be downgraded
	if !db.IsReadOnly() && (dbVersion == nil || (dbVersion.LE(*c.DBCheckpointVersion))) {
		if err := dv.SetDBVersion(db, c.DBCheckpointVersion); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// checkDBVersion checks whether need to verify or reset the DB version
func checkDBVersion(c dbCheckConfig, dbVersion *semver.Version) (dbAction, error) {
	// If the saved DB version is higher than the app version, abort.
	// Otherwise DB corruption could occur.
	if dbVersion != nil && dbVersion.GT(*c.DBCheckpointVersion) {
		return doNothing, fmt.Errorf("Cannot use newer DB version=%v with older check point version=%v", dbVersion, c.DBCheckpointVersion)
	}

	// Verify the DB if the version detection says to, or if it was requested on the command line
	if shouldVerifyDB(dbVersion, c.DBCheckpointVersion) || c.ForceVerify {
		if c.ResetCorruptDB {
			return doResetCorruptDB, nil
		}

		return doCheckDB, nil
	}

	return doNothing, nil
}

func shouldVerifyDB(dbVersion, checkpointVersion *semver.Version) bool {
	// If the dbVersion is not set, verify
	if dbVersion == nil {
		return true
	}

	// If the dbVersion is less than the verification checkpoint version,
	// verify
	if dbVersion.LT(*checkpointVersion) {
		return true
	}

	return false
}
