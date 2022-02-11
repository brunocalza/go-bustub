//package concurrency
//package transaction
package access

import (
	"github.com/ryogrid/SamehadaDB/common"
	"github.com/ryogrid/SamehadaDB/recovery"
	"github.com/ryogrid/SamehadaDB/types"
)

type TransactionManager struct {
	// TODO: (SDB) must ensure atomicity
	next_txn_id types.TxnID
	// lock_manager *LockManager //__attribute__((__unused__))
	log_manager *recovery.LogManager // __attribute__((__unused__))
	// /** The global transaction latch is used for checkpointing. */
	global_txn_latch common.ReaderWriterLatch
}

var txn_map map[types.TxnID]*Transaction = make(map[types.TxnID]*Transaction)

func NewTransactionManager(log_manager *recovery.LogManager) *TransactionManager {
	return &TransactionManager{0, log_manager, common.NewRWLatch()}
}

func (transaction_manager *TransactionManager) Begin(txn *Transaction) *Transaction {
	// Acquire the global transaction latch in shared mode.
	transaction_manager.global_txn_latch.RLock()
	var txn_ret *Transaction = txn

	if txn_ret == nil {
		transaction_manager.next_txn_id += 1
		txn_ret = NewTransaction(transaction_manager.next_txn_id)
	}

	if common.EnableLogging {
		log_record := recovery.NewLogRecordTxn(txn_ret.GetTransactionId(), txn_ret.GetPrevLSN(), recovery.BEGIN)
		lsn := transaction_manager.log_manager.AppendLogRecord(log_record)
		txn_ret.SetPrevLSN(lsn)
	}

	txn_map[txn_ret.GetTransactionId()] = txn_ret
	return txn_ret
}

func (transaction_manager *TransactionManager) Commit(txn *Transaction) {
	txn.SetState(COMMITTED)

	// TODO: (SDB) not ported yet
	/*
	   // Perform all deletes before we commit.
	   auto write_set = txn.GetWriteSet()
	   while (!write_set.empty()) {
	     auto &item = write_set.back()
	     auto table = item.table;
	     if (item.wtype == WType::DELETE) {
	       // Note that this also releases the lock when holding the page latch.
	       table.ApplyDelete(item.rid, txn)
	     }
	     write_set.pop_back()
	   }
	   write_set.clear()
	*/

	if common.EnableLogging {
		log_record := recovery.NewLogRecordTxn(txn.GetTransactionId(), txn.GetPrevLSN(), recovery.COMMIT)
		lsn := transaction_manager.log_manager.AppendLogRecord(log_record)
		txn.SetPrevLSN(lsn)
		transaction_manager.log_manager.Flush()
	}

	// Release all the locks.
	transaction_manager.releaseLocks(txn)
	// Release the global transaction latch.
	transaction_manager.global_txn_latch.RUnlock()
}

func (transaction_manager *TransactionManager) Abort(txn *Transaction) {
	txn.SetState(ABORTED)

	// TODO: (SDB) not ported yet
	/*
	   // Rollback before releasing the access.
	   auto write_set = txn.GetWriteSet();
	   while (!write_set.empty()) {
	     auto &item = write_set.back()
	     auto table = item.table
	     if (item.wtype == WType::DELETE) {
	       table.RollbackDelete(item.rid_, txn)
	     } else if (item.wtype == WType::INSERT) {
	       // Note that this also releases the lock when holding the page latch.
	       table.ApplyDelete(item.rid, txn)
	     } else if (item.wtype_ == WType::UPDATE) {
	       table.UpdateTuple(item.tuple, item.rid_, txn)
	     }
	     write_set.pop_back()
	   }
	   write_set.clear()
	*/

	if common.EnableLogging {
		log_record := recovery.NewLogRecordTxn(txn.GetTransactionId(), txn.GetPrevLSN(), recovery.ABORT)
		lsn := transaction_manager.log_manager.AppendLogRecord(log_record)
		txn.SetPrevLSN(lsn)
	}

	// Release all the locks.
	transaction_manager.releaseLocks(txn)
	// Release the global transaction latch.
	transaction_manager.global_txn_latch.RUnlock()
}

func (transaction_manager *TransactionManager) BlockAllTransactions() {
	transaction_manager.global_txn_latch.WLock()
}

func (transaction_manager *TransactionManager) ResumeTransactions() {
	transaction_manager.global_txn_latch.WUnlock()
}

func (transaction_manager *TransactionManager) releaseLocks(txn *Transaction) {
	// TODO: (SDB) not ported yet
	/*
	   	var lock_set : unordered_set<RID>
	       for (item : *txn.GetExclusiveLockSet()) {
	         lock_set.emplace(item)
	       }
	       for (item : *txn.GetSharedLockSet()) {
	         lock_set.emplace(item)
	       }
	       for (locked_rid : lock_set) {
	         lock_manager.Unlock(txn, locked_rid)
	       }
	*/
}