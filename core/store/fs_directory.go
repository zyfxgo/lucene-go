package store

// FSDirectory Base class for Directory implementations that store index files in the file system.
// There are currently three core subclasses:
// SimpleFSDirectory is a straightforward implementation using Files.newByteChannel. However, it has poor concurrent performance (multiple threads will bottleneck) as it synchronizes when multiple threads read from the same file.
// NIOFSDirectory uses java.nio's FileChannel's positional io when reading to avoid synchronization when reading from the same file. Unfortunately, due to a Windows-only Sun JRE bug  this is a poor choice for Windows, but on all other platforms this is the preferred choice. Applications using Thread.interrupt() or Future.cancel(boolean) should use RAFDirectory instead. See NIOFSDirectory java doc for details.
// MMapDirectory uses memory-mapped IO when reading. This is a good choice if you have plenty of virtual memory relative to your index size, eg if you are running on a 64 bit JRE, or you are running on a 32 bit JRE but your index sizes are small enough to fit into the virtual memory space. Java has currently the limitation of not being able to unmap files from user code. The files are unmapped, when GC releases the byte buffers. Due to this bug  in Sun's JRE, MMapDirectory's IndexInput.close is unable to close the underlying OS file handle. Only when GC finally collects the underlying objects, which could be quite some time later, will the file handle be closed. This will consume additional transient disk usage: on Windows, attempts to delete or overwrite the files will result in an exception; on other platforms, which typically have a "delete on last close" semantics, while such operations will succeed, the bytes are still consuming space on disk. For many applications this limitation is not a problem (e.g. if you have plenty of disk space, and you don't rely on overwriting files on Windows) but it's still an important limitation to be aware of. This class supplies a (possibly dangerous) workaround mentioned in the bug report, which may fail on non-Sun JVMs.
// Unfortunately, because of system peculiarities, there is no single overall best implementation. Therefore, we've added the open method, to allow Lucene to choose the best FSDirectory implementation given your environment, and the known limitations of each implementation. For users who have no reason to prefer a specific implementation, it's best to simply use open. For all others, you should instantiate the desired implementation directly.
// NOTE: Accessing one of the above subclasses either directly or indirectly from a thread while it's interrupted can close the underlying channel immediately if at the same time the thread is blocked on IO. The channel will remain closed and subsequent access to the index will throw a ClosedChannelException. Applications using Thread.interrupt() or Future.cancel(boolean) should use the slower legacy RAFDirectory from the misc Lucene module instead.
// The locking implementation is by default NativeFSLockFactory, but can be changed by passing in a custom LockFactory instance.
// See Also: Directory
type FSDirectory interface {
	BaseDirectory

	// GetDirectory Returns: the underlying filesystem directory
	GetDirectory() string
}
