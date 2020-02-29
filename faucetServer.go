package main

import (
"crypto/ecdsa"
"crypto/sha256"
"encoding/json"
"flag"
"fmt"
"math/big"
"strconv"
"strings"

"github.com/unichainplatform/unichain/accountmanager"
"github.com/unichainplatform/unichain/common"
"github.com/unichainplatform/unichain/crypto"
"github.com/unichainplatform/unichain/types"
"github.com/unichainplatform/unichain/utils/rlp"

"github.com/syndtr/goleveldb/leveldb"
"github.com/syndtr/goleveldb/leveldb/errors"
"github.com/syndtr/goleveldb/leveldb/filter"
"github.com/syndtr/goleveldb/leveldb/opt"

tc "github.com/unichainplatform/unichain/test/common"
"net/http"
)

var (
	gaslimit = uint64(20000000)
)

type GenAction struct {
	*types.Action
	PrivateKey *ecdsa.PrivateKey
}

func createAccount(accountName, from common.Name, nonce uint64, publickey common.PubKey, prikey *ecdsa.PrivateKey, chain_id int) (error, common.Hash) {
	account := &accountmanager.CreateAccountAction{
		AccountName: accountName,
		Founder:     common.Name(""),
		PublicKey:   publickey,
		Description: "create by unichain wallet",
	}
	payload, err := rlp.EncodeToBytes(account)
	if err != nil {
		return fmt.Errorf("rlp payload err %v", err), common.Hash{}
	}
	gc := newGeAction(types.CreateAccount, from, "unichain.account", nonce, 0, gaslimit, nil, payload, prikey)
	var gcs []*GenAction
	gcs = append(gcs, gc)
	return sendTxTest(gcs, chain_id)
}

func GeneragePubKey() (common.PubKey, *ecdsa.PrivateKey) {
	prikey, _ := crypto.GenerateKey()
	return common.BytesToPubKey(crypto.FromECDSAPub(&prikey.PublicKey)), prikey
}

func newGeAction(at types.ActionType, from, to common.Name, nonce uint64, assetid uint64, gaslimit uint64, amount *big.Int, payload []byte, prikey *ecdsa.PrivateKey) *GenAction {
	action := types.NewAction(at, from, to, nonce, assetid, gaslimit, amount, payload, nil)

	return &GenAction{
		Action:     action,
		PrivateKey: prikey,
	}
}

func sendTxTest(gcs []*GenAction, chain_id int) (error, common.Hash) {
	signer := types.NewSigner(big.NewInt(int64(chain_id)))
	var actions []*types.Action
	for _, v := range gcs {
		actions = append(actions, v.Action)
	}
	tx := types.NewTransaction(uint64(0), big.NewInt(1000000000), actions...)
	for _, v := range gcs {
		keypair := types.MakeKeyPair(v.PrivateKey, []uint64{0})
		err := types.SignActionWithMultiKey(v.Action, tx, signer, 0, []*types.KeyPair{keypair})
		if err != nil {
			return fmt.Errorf("SignAction err %v", err), common.Hash{}
		}

	}
	rawtx, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return err, common.Hash{}
	}
	hash, err := tc.SendRawTx(rawtx)
	if err != nil {
		return err, common.Hash{}
	}
	fmt.Printf("hash: %x", hash)
	return nil, hash
}

type RespForm struct {
	Code int    `json:"code"`
	Msg  string `json:"msg,omitempty"`
}

type DbRecord struct {
	Count uint `json:"count"`
}

type DbStatus struct {
	Pos int `json:"pos"`
}

var db_status_key = []byte("db_status_key")
var pn = flag.String("pn", "walletservice.u", "user name")
var pk = flag.String("pk", "", "priv key")
var climit = flag.String("l", "5", "create limit per user")
var host = flag.String("h", "47.115.149.93", "ip")

func main() {
	flag.Parse()

	na := *pn
	pri := *pk
	prikey, _ := crypto.HexToECDSA(pri)

	cl, _ := strconv.Atoi(*climit)
	url := "http://" + *host + ":8080"

	var chain_id int
	chain_id = 100

	fmt.Printf("user_name:%v priv_key:%v climit:%v \n", na, pri, cl)
	fmt.Printf("url:%v chain_id:%v \n", url, chain_id)

	// level db
	db_path := "./ldb/"

	//os.RemoveAll(db_path)
	db, err := leveldb.OpenFile(db_path, &opt.Options{
		OpenFilesCacheCapacity: 16,
		BlockCacheCapacity:     8 * opt.MiB,
		WriteBuffer:            4 * opt.MiB,
		Filter:                 filter.NewBloomFilter(100),
	})
	if _, corrupted := err.(*errors.ErrCorrupted); corrupted {
		db, err = leveldb.RecoverFile(db_path, nil)
	}
	defer db.Close()

	http.HandleFunc("/wallet_account_creation", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		resform := &RespForm{
			Code: 200,
		}

		// this is a for-break imp
		for {
			var accname string
			var pubkey string
			var deviceid string

			if val, ok := q["accname"]; ok {
				accname = val[0]
			} else {
				resform.Code = 400
				resform.Msg = "accname miss!"
				break
			}

			if val, ok := q["pubkey"]; ok {
				pubkey = val[0]
			} else {
				resform.Code = 400
				resform.Msg = "pubkey miss!"
				break
			}

			if val, ok := q["deviceid"]; ok {
				deviceid = val[0]
			} else {
				resform.Code = 400
				resform.Msg = "deviceid miss!"
				break
			}

			// X-Real-IP from header
			var remote_addr string
			var ip_str string
			if x_real_ip, ok := r.Header["X-Forwarded-For"]; ok {
				//do something here
				ip_str= x_real_ip[0]
			}else{
				remote_addr = r.RemoteAddr
				idx := strings.Index(remote_addr,":")
				ip_str=remote_addr[:idx]
			}
			//fmt.Printf("%v\n",r.Header)

			// ip limit check
			db_key := sha256.Sum256([]byte(ip_str))

			db_record := DbRecord{}
			if db_value, err := db.Get(db_key[:], nil); err != nil {
				if err != errors.ErrNotFound {
					resform.Code = 500
					resform.Msg = err.Error()
					break
				}
			} else {
				json.Unmarshal(db_value, &db_record)
			}
			fmt.Printf("db_r:%v\n", db_record)

			// output log
			fmt.Printf("ip=%s&accname=%s&pukkey=%s&deviceid=%s\n",
				ip_str, accname, pubkey, deviceid)

			// max create count
			if db_record.Count > uint(cl) {
				resform.Code = 404
				resform.Msg = "exceed max count"
				fmt.Println("exceed max count")
				break
			}

			// fetch postion for sender
			db_status := DbStatus{Pos: 1}
			if db_value, err := db.Get(db_status_key, nil); err != nil {
				if err != errors.ErrNotFound {
					resform.Code = 500
					resform.Msg = err.Error()
					break
				}
			} else {
				json.Unmarshal(db_value, &db_status)
			}
			// sync pos
			send_pos := db_status.Pos
			if 9 == db_status.Pos {
				db_status.Pos = 1
			} else {
				db_status.Pos += 1
			}
			db_value, _ := json.Marshal(db_status)
			db.Put(db_status_key, db_value[:], nil)

			// rpc create account
			tc.SetDefultURL(url)
			sender_na := common.Name(na + strconv.Itoa(send_pos))
			fmt.Println("sender_na:", sender_na)
			cn, _ := tc.GetNonce(sender_na)

			if err, hash := createAccount(common.Name(accname), sender_na, cn,
				common.HexToPubKey(pubkey), prikey, chain_id); err != nil {
				resform.Code = 500
				resform.Msg = err.Error()
				break
			} else {
				resform.Msg = hash.String()
				// count ip limit
				db_record.Count += 1
				db_value, _ := json.Marshal(db_record)
				// save back db record
				db.Put(db_key[:], db_value[:], nil)

				fmt.Printf("trans sent, count:%v\n", db_record.Count)
				break
			}

			// break at the end
			//break
		}

		//http.NotFound(w,r)
		b, err := json.Marshal(resform)
		if err != nil {
			http.Error(w, err.Error(), 502)
		} else {
			// everything ok return json
			w.Header().Set("Content-Type", "application/json;charset=UTF-8")
			w.Header().Set("Access-Control-Allow-Origin", "*")
			if resform.Code != 200 {
				http.Error(w, string(b), resform.Code)
			} else {
				w.Write(b)
			}
		}
	})

	//fmt.Println("listen and serve")
	http.ListenAndServe(":12345", nil)
}