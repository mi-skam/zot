//go:build goolm

package matrix

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.mau.fi/util/dbutil"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"

	_ "maunium.net/go/mautrix/crypto/goolm" // pure-Go olm — keeps the build CGO-free
	_ "modernc.org/sqlite"                  // CGO-free sqlite driver, registers "sqlite"
)

// EnableCrypto initialises the olm/megolm machine backed by a sqlite
// store at $ZOT_HOME/matrix-crypto/store.db. After a successful call
// the client transparently decrypts inbound m.room.encrypted events
// (re-dispatched as event.EventMessage) and encrypts outbound sends
// to encrypted rooms. Passphrase pickles the key store.
//
// Returns a closer the caller must Close on shutdown. Callers must
// only invoke this when a passphrase is configured; an empty
// passphrase is a hard error, not a silent no-crypto fallback.
//
// This implementation is compiled only with the "goolm" build tag so
// the default CGO-free build stays free of the olm crypto store.
// Build with: CGO_ENABLED=0 go build -tags goolm ./...
func EnableCrypto(ctx context.Context, cli *mautrix.Client, zotHome, passphrase string) (io.Closer, error) {
	if passphrase == "" {
		return nil, fmt.Errorf("crypto passphrase is empty")
	}
	dbPath := CryptoStorePath(zotHome)
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return nil, err
	}
	rawDB, err := sql.Open("sqlite", "file:"+dbPath+"?_pragma=busy_timeout(10000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	db, err := dbutil.NewWithDB(rawDB, "sqlite3")
	if err != nil {
		return nil, err
	}
	helper, err := cryptohelper.NewCryptoHelper(cli, []byte(passphrase), db)
	if err != nil {
		return nil, err
	}
	if err := helper.Init(ctx); err != nil {
		return nil, fmt.Errorf("init crypto: %w", err)
	}
	// helper.Init sets cli.Crypto; sanity-check because a nil Crypto
	// silently downgrades to plaintext-only.
	if cli.Crypto == nil {
		return nil, fmt.Errorf("crypto helper did not attach to client")
	}
	return helper, nil
}
