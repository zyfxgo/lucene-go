package index

import (
	"github.com/geange/lucene-go/core/types"
	"go.uber.org/atomic"
)

// DocumentsWriter This class accepts multiple added documents and directly writes segment files.
// Each added document is passed to the indexing chain, which in turn processes the document into
// the different codec formats. Some formats write bytes to files immediately, e.g. stored fields
// and term vectors, while others are buffered by the indexing chain and written only on flush.
// Once we have used our allowed RAM buffer, or the number of added docs is large enough (in the
// case we are flushing by doc count instead of RAM usage), we create a real segment and flush it
// to the Directory. Threads: Multiple threads are allowed into addDocument at once. There is an
// initial synchronized call to DocumentsWriterFlushControl.obtainAndLock() which allocates a DWPT
// for this indexing thread. The same thread will not necessarily get the same DWPT over time.
// Then updateDocuments is called on that DWPT without synchronization (most of the "heavy lifting"
// is in this call). Once a DWPT fills up enough RAM or hold enough documents in memory the DWPT
// is checked out for flush and all changes are written to the directory. Each DWPT corresponds to
// one segment being written. When flush is called by IndexWriter we check out all DWPTs that are
// associated with the current DocumentsWriterDeleteQueue out of the DocumentsWriterPerThreadPool
// and write them to disk. The flush process can piggy-back on incoming indexing threads or even
// block them from adding documents if flushing can't keep up with new documents being added.
// Unless the stall control kicks in to block indexing threads flushes are happening concurrently
// to actual index requests. Exceptions: Because this class directly updates in-memory posting lists,
// and flushes stored fields and term vectors directly to files in the directory, there are certain
// limited times when an exception can corrupt this state. For example, a disk full while flushing
// stored fields leaves this file in a corrupt state. Or, an OOM exception while appending to the
// in-memory posting lists can corrupt that posting list. We call such exceptions "aborting exceptions".
// In these cases we must call abort() to discard all docs added since the last flush. All other
// exceptions ("non-aborting exceptions") can still partially update the index structures.
// These updates are consistent, but, they represent only a part of the document seen up until the
// exception was hit. When this happens, we immediately mark the document as deleted so that the
// document is always atomically ("all or none") added to the index.
type DocumentsWriter struct {
	pendingNumDocs     *atomic.Int64
	flushNotifications FlushNotifications
}

func (d *DocumentsWriter) updateDocuments(docs []types.IndexableFieldIterator, delNode Node) (int64, error) {
	panic("")
}

type FlushNotifications interface {
	// DeleteUnusedFiles Called when files were written to disk that are not used anymore. It's the implementation's
	// responsibility to clean these files up
	DeleteUnusedFiles(files map[string]struct{})

	// FlushFailed Called when a segment failed to flush.
	FlushFailed(info *SegmentInfo)

	// AfterSegmentsFlushed Called after one or more segments were flushed to disk.
	AfterSegmentsFlushed() error

	// Should be called if a flush or an indexing operation caused a tragic / unrecoverable event.

	// OnDeletesApplied Called once deletes have been applied either after a flush or on a deletes call
	OnDeletesApplied()

	// OnTicketBacklog Called once the DocumentsWriter ticket queue has a backlog. This means there
	// is an inner thread that tries to publish flushed segments but can't keep up with the other
	// threads flushing new segments. This likely requires other thread to forcefully purge the buffer
	// to help publishing. This can't be done in-place since we might hold index writer locks when
	// this is called. The caller must ensure that the purge happens without an index writer lock being held.
	// See Also: purgeFlushTickets(boolean, IOUtils.IOConsumer)
	OnTicketBacklog()
}
