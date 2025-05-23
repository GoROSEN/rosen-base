package blockchain

import (
	"context"
	"errors"
	"math/big"
	"strconv"
	"time"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	solana "github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	confirm "github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/google/martian/log"
	"golang.org/x/time/rate"
)

type SolanaChainAccess struct {
	rpcClient *rpc.Client
	cfg       *config.BlockchainConfig
}

func NewSolanaChainAccess(chainCfg *config.BlockchainConfig) (BlockChainAccess, error) {

	var client1 *rpc.Client
	if chainCfg.RateLimit > 0 {
		client1 = rpc.NewWithCustomRPCClient(rpc.NewWithLimiter(
			chainCfg.Endpoint,
			rate.Every(time.Second), // time frame
			chainCfg.RateLimit,      // limit of requests per time frame
		))
	} else {
		client1 = rpc.New(chainCfg.Endpoint)
	}

	return &SolanaChainAccess{client1, chainCfg}, nil
}

func (sca *SolanaChainAccess) NewWallet() (string, string) {
	wa, _ := solana.NewRandomPrivateKey()
	return wa.PublicKey().String(), wa.String()
}

func (sca *SolanaChainAccess) NewTokenAccount(mintAddress string, walletAddress string) (string, error) {

	feePayer, err := solana.PrivateKeyFromBase58(sca.cfg.Funder)
	if err != nil {
		log.Errorf("failed to get fee payer, err: %v", err)
		return "", err
	}
	toAddr, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		log.Errorf("failed to get mint account, err: %v", err)
		return "", err
	}
	mintAddr, err := solana.PublicKeyFromBase58(mintAddress)
	if err != nil {
		log.Errorf("failed to get mint account, err: %v", err)
		return "", err
	}

	ctx := context.TODO()
	instruction0 := associatedtokenaccount.NewCreateInstruction(feePayer.PublicKey(), toAddr, mintAddr).Build()
	recent, err := sca.rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Errorf("get recent blockhash failed: %v", err)
		return "", err
	}

	tx, err := solana.NewTransaction([]solana.Instruction{instruction0}, recent.Value.Blockhash, solana.TransactionPayer(feePayer.PublicKey()))
	if err != nil {
		log.Errorf("create transaction failed: %v", err)
		return "", err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if feePayer.PublicKey().Equals(key) {
				return &feePayer
			}
			return nil
		},
	)
	if err != nil {
		log.Errorf("sign transaction failed: %v", err)
		return "", err
	}
	// Pretty print the transaction:
	// tx.EncodeTree(text.NewTreeEncoder(os.Stdout, "Transfer USDT"))

	sig, err := sca.rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		log.Errorf("send transaction failed: %v", err)
		return "", err
	}
	return sig.String(), nil
}

func (sca *SolanaChainAccess) QueryCoin(address string) (*big.Int, error) {
	// get account
	account, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		log.Errorf("cannot parse %v as public key", address)
		return nil, err
	}
	// get balance
	if out, err := sca.rpcClient.GetBalance(context.TODO(),
		account,
		rpc.CommitmentFinalized,
	); err != nil {
		return nil, err
	} else {
		// log.Infof("Coin Amount = %v", out.Value)
		return big.NewInt(int64(out.Value)), nil
	}
}

func (sca *SolanaChainAccess) QueryToken(address string, contractAddress string) (*big.Int, error) {
	// get token account
	tokenAccount, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		log.Errorf("cannot parse %v as public key", address)
		return nil, err
	}
	// get balance
	if out, err := sca.rpcClient.GetTokenAccountBalance(context.TODO(),
		tokenAccount,
		rpc.CommitmentFinalized,
	); err != nil {
		return nil, err
	} else {
		// log.Infof("Token Amount = %v", out.Value.Amount)
		if amount, err := strconv.ParseInt(out.Value.Amount, 10, 64); err != nil {
			log.Errorf("invalid amount %v", err)
			return nil, err
		} else {
			return big.NewInt(amount), nil
		}
	}
}

func (sca *SolanaChainAccess) TransferCoin(from, to string, value *big.Int) (string, error) {
	feePayer, err := solana.PrivateKeyFromBase58(sca.cfg.Funder)
	if err != nil {
		log.Errorf("failed to get fee payer, err: %v", err)
		return "", err
	}

	fromPrikey, err := solana.PrivateKeyFromBase58(from)
	if err != nil {
		log.Errorf("failed to get from account, err: %v", err)
		return "", err
	}

	toAddr, err := solana.PublicKeyFromBase58(to)
	if err != nil {
		log.Errorf("failed to get mint account, err: %v", err)
		return "", err
	}
	ctx := context.TODO()
	instruction0 := token.NewTransferInstruction(
		value.Uint64(),
		fromPrikey.PublicKey(),
		toAddr,
		fromPrikey.PublicKey(),
		[]solana.PublicKey{feePayer.PublicKey(), fromPrikey.PublicKey()},
	).Build()
	recent, err := sca.rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Errorf("get recent blockhash failed: %v", err)
		return "", err
	}

	tx, err := solana.NewTransaction([]solana.Instruction{instruction0}, recent.Value.Blockhash, solana.TransactionPayer(feePayer.PublicKey()))
	if err != nil {
		log.Errorf("create transaction failed: %v", err)
		return "", err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if feePayer.PublicKey().Equals(key) {
				return &feePayer
			}
			if fromPrikey.PublicKey().Equals(key) {
				return &fromPrikey
			}
			return nil
		},
	)
	if err != nil {
		log.Errorf("sign transaction failed: %v", err)
		return "", err
	}
	// Pretty print the transaction:
	// tx.EncodeTree(text.NewTreeEncoder(os.Stdout, "Transfer USDT"))

	sig, err := sca.rpcClient.SendTransaction(ctx, tx)
	if err != nil {
		log.Errorf("send transaction failed: %v", err)
		return "", err
	}
	return sig.String(), nil
}

func (sca *SolanaChainAccess) TransferToken(from, to string, value *big.Int, contractAddress string, decimals uint64) (string, error) {

	feePayer, err := solana.PrivateKeyFromBase58(sca.cfg.Funder)
	if err != nil {
		log.Errorf("failed to get fee payer, err: %v", err)
		return "", err
	}

	fromPrikey, err := solana.PrivateKeyFromBase58(from)
	if err != nil {
		log.Errorf("failed to get from account, err: %v", err)
		return "", err
	}

	mintAddr, err := solana.PublicKeyFromBase58(contractAddress)
	if err != nil {
		log.Errorf("failed to get mint account, err: %v", err)
		return "", err
	}

	fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(fromPrikey.PublicKey(), mintAddr)
	if err != nil {
		log.Errorf("cannot get from token ATA: %v", err)
		return "", err
	} else {
		// 检查from账户余额是否充足，嫌慢可取消，但是要多烧gas
		if out, err := sca.rpcClient.GetTokenAccountBalance(context.TODO(),
			fromTokenAccount,
			rpc.CommitmentFinalized,
		); err != nil {
			return "cannot get from account balance", err
		} else {
			// log.Infof("Token Amount = %v", out.Value.Amount)
			if amount, err := strconv.ParseInt(out.Value.Amount, 10, 64); err != nil {
				log.Errorf("invalid from token amount %v", err)
				return "invalid from token amount", err
			} else if value.Cmp(big.NewInt(amount)) > 0 {
				log.Errorf("insufficient from token")
				return "insufficient token", errors.New("insufficient token")
			}
		}
	}
	toAddr, err := solana.PublicKeyFromBase58(to)
	if err != nil {
		log.Errorf("failed to get mint account, err: %v", err)
		return "", err
	}
	instructions := []solana.Instruction{}
	// caculate fees
	microlamports, limits, _ := sca.fillFees()
	feesInit := computebudget.NewSetComputeUnitPriceInstructionBuilder()
	feesInit.SetMicroLamports(microlamports)
	feesIx := feesInit.Build()
	instructions = append(instructions, feesIx)
	feesLimit := computebudget.NewSetComputeUnitLimitInstructionBuilder()
	feesLimit.SetUnits(uint32(limits))
	limitIx := feesLimit.Build()
	instructions = append(instructions, limitIx)

	// find accounts
	toTokenAccount, _, err := solana.FindAssociatedTokenAddress(toAddr, mintAddr)
	if err != nil {
		log.Errorf("cannot get token ATA: %v", err)
		return "", err
	} else {
		if _, err := sca.rpcClient.GetTokenAccountBalance(context.TODO(),
			toTokenAccount,
			rpc.CommitmentFinalized,
		); err != nil {
			log.Errorf("cannot get to account balance, %v", err)
		}
		if result, err := sca.rpcClient.GetAccountInfo(context.TODO(), toTokenAccount); err != nil {
			log.Errorf("cannot get to account(ata) info, %v", err)
			log.Infof("creating ata for %v in %v", to, contractAddress)
			instruction0 := associatedtokenaccount.NewCreateInstruction(feePayer.PublicKey(), toAddr, mintAddr).Build()
			instructions = append(instructions, instruction0)
		} else {
			log.Infof("get ata account info: %v", result.Value)
		}
	}
	var signers []solana.PublicKey
	if feePayer.PublicKey() == fromPrikey.PublicKey() {
		signers = []solana.PublicKey{feePayer.PublicKey(), fromPrikey.PublicKey()}
	} else {
		signers = []solana.PublicKey{feePayer.PublicKey()}
	}
	instruction0 := token.NewTransferCheckedInstruction(
		value.Uint64(),
		uint8(decimals),
		fromTokenAccount,
		mintAddr,
		toTokenAccount,
		fromPrikey.PublicKey(),
		signers,
	).Build()
	instructions = append(instructions, instruction0)
	ctx := context.TODO()
	recent, err := sca.rpcClient.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Errorf("get recent blockhash failed: %v", err)
		return "", err
	}

	tx, err := solana.NewTransaction(instructions, recent.Value.Blockhash, solana.TransactionPayer(feePayer.PublicKey()))
	if err != nil {
		log.Errorf("create transaction failed: %v", err)
		return "", err
	}

	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if feePayer.PublicKey().Equals(key) {
				return &feePayer
			}
			if fromPrikey.PublicKey().Equals(key) {
				return &fromPrikey
			}
			return nil
		},
	)
	if err != nil {
		log.Errorf("sign transaction failed: %v", err)
		return "", err
	}
	// Pretty print the transaction:
	// tx.EncodeTree(text.NewTreeEncoder(os.Stdout, "Transfer SPL"))
	var sig solana.Signature
	wsClient, err := ws.Connect(context.Background(), sca.cfg.WsEndpoint)
	if err != nil {
		sig, err = sca.rpcClient.SendTransaction(ctx, tx)
		if err != nil {
			log.Errorf("send transaction via rpc failed: %v", err)
			return "", err
		}
	} else {
		sig, err = confirm.SendAndConfirmTransaction(ctx, sca.rpcClient, wsClient, tx)
		if err != nil {
			log.Errorf("send transaction via wss rpc failed: %v", err)
			return "", err
		}
	}
	return sig.String(), nil
}

func (sca *SolanaChainAccess) MintNFT(to string, contractAddress string, tokenId uint64, tokenUri string) (string, error) {
	return "", nil
}

func (sca *SolanaChainAccess) TransferNFT(from, to string, contractAddress string, tokenId uint64) (string, error) {
	return "", nil
}

func (sca *SolanaChainAccess) FindTokenAccount(contractAddress string, walletAddress string) (string, error) {

	fromPubkey, err := solana.PublicKeyFromBase58(walletAddress)
	if err != nil {
		log.Errorf("failed to get from account, err: %v", err)
		return "", err
	}

	mintAddr, err := solana.PublicKeyFromBase58(contractAddress)
	if err != nil {
		log.Errorf("failed to get mint account, err: %v", err)
		return "", err
	}

	fromTokenAccount, _, err := solana.FindAssociatedTokenAddress(fromPubkey, mintAddr)
	if err != nil {
		log.Errorf("cannot get from token ATA: %v", err)
		return "", err
	}
	return fromTokenAccount.String(), nil
}

func (sca *SolanaChainAccess) ConfirmTransaction(txhash string) (bool, error) {
	ctx := context.TODO()
	sig, err := solana.SignatureFromBase58(txhash)
	if err != nil {
		log.Errorf("cannot get signature from %v: %v", txhash, err)
		return false, err
	}
	result, err := sca.rpcClient.GetTransaction(ctx, sig, nil)
	if err != nil {
		if err == rpc.ErrNotFound {
			return false, nil
		}
		log.Errorf("get transaction error for %v: %v", txhash, err)
		return false, err
	}
	if result == nil {
		return false, nil
	}
	return true, nil
}

func (sca *SolanaChainAccess) fillFees() ( /*microlamports*/ uint64 /*feesLimit*/, uint64, error) {
	ctx := context.TODO()
	result, err := sca.rpcClient.GetFees(ctx, rpc.CommitmentConfirmed)
	if err != nil {
		return uint64(sca.cfg.GasPrice), uint64(sca.cfg.GasLimit), err
	}
	return result.Value.FeeCalculator.LamportsPerSignature, uint64(sca.cfg.GasLimit), nil
}
