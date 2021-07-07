package buffer

import (
	"errors"

	"github.com/brunocalza/go-bustub/storage/disk"
	"github.com/brunocalza/go-bustub/storage/page"
)

//MaxPoolSize is the size of the pool
const MaxPoolSize = 1000

//BufferPoolManager represents the buffer pool manager
type BufferPoolManager struct {
	diskManager disk.DiskManager
	pages       [MaxPoolSize]*page.Page
	replacer    *ClockReplacer
	freeList    []FrameID
	pageTable   map[page.PageID]FrameID
}

// FetchPage fetches the requested page from the buffer pool.
func (b *BufferPoolManager) FetchPage(pageID page.PageID) *page.Page {
	// if it is on buffer pool return it
	if frameID, ok := b.pageTable[pageID]; ok {
		pg := b.pages[frameID]
		pg.IncPinCount()
		(*b.replacer).Pin(frameID)
		return pg
	}

	// get the id from free list or from replacer
	frameID, isFromFreeList := b.getFrameID()
	if frameID == nil {
		return nil
	}

	if !isFromFreeList {
		// remove page from current frame
		currentPage := b.pages[*frameID]
		if currentPage != nil {
			if currentPage.IsDirty() {
				data := currentPage.Data()
				b.diskManager.WritePage(currentPage.ID(), data[:])
			}

			delete(b.pageTable, currentPage.ID())
		}
	}

	data := make([]byte, page.PageSize)
	err := b.diskManager.ReadPage(pageID, data)
	if err != nil {
		return nil
	}
	var pageData [page.PageSize]byte
	copy(pageData[:], data)
	pg := page.New(pageID, 1, false, &pageData)
	b.pageTable[pageID] = *frameID
	b.pages[*frameID] = pg

	return pg
}

// UnpinPage unpins the target page from the buffer pool.
func (b *BufferPoolManager) UnpinPage(pageID page.PageID, isDirty bool) error {
	if frameID, ok := b.pageTable[pageID]; ok {
		pg := b.pages[frameID]
		pg.DecPinCount()

		if pg.PinCount() <= 0 {
			(*b.replacer).Unpin(frameID)
		}

		if pg.IsDirty() || isDirty {
			pg.SetIsDirty(true)
		} else {
			pg.SetIsDirty(false)
		}

		return nil
	}

	return errors.New("could not find page")
}

// FlushPage Flushes the target page to disk.
func (b *BufferPoolManager) FlushPage(pageID page.PageID) bool {
	if frameID, ok := b.pageTable[pageID]; ok {
		pg := b.pages[frameID]
		pg.DecPinCount()

		data := pg.Data()
		b.diskManager.WritePage(pageID, data[:])
		pg.SetIsDirty(false)

		return true
	}

	return false
}

// NewPage allocates a new page in the buffer pool with the disk manager help
func (b *BufferPoolManager) NewPage() *page.Page {
	frameID, isFromFreeList := b.getFrameID()
	if frameID == nil {
		return nil
	}

	if !isFromFreeList {
		// remove page from current frame
		currentPage := b.pages[*frameID]
		if currentPage != nil {
			if currentPage.IsDirty() {
				data := currentPage.Data()
				b.diskManager.WritePage(currentPage.ID(), data[:])
			}

			delete(b.pageTable, currentPage.ID())
		}
	}

	// allocates new page
	pageID := b.diskManager.AllocatePage()
	pg := page.NewEmpty(pageID)

	b.pageTable[pageID] = *frameID
	b.pages[*frameID] = pg

	return pg
}

// DeletePage deletes a page from the buffer pool.
func (b *BufferPoolManager) DeletePage(pageID page.PageID) error {
	var frameID FrameID
	var ok bool
	if frameID, ok = b.pageTable[pageID]; !ok {
		return nil
	}

	page := b.pages[frameID]

	if page.PinCount() > 0 {
		return errors.New("Pin count greater than 0")
	}
	delete(b.pageTable, page.ID())
	(*b.replacer).Pin(frameID)
	b.diskManager.DeallocatePage(pageID)

	b.freeList = append(b.freeList, frameID)

	return nil

}

// FlushAllpages flushes all the pages in the buffer pool to disk.
func (b *BufferPoolManager) FlushAllpages() {
	for pageID := range b.pageTable {
		b.FlushPage(pageID)
	}
}

func (b *BufferPoolManager) getFrameID() (*FrameID, bool) {
	if len(b.freeList) > 0 {
		frameID, newFreeList := b.freeList[0], b.freeList[1:]
		b.freeList = newFreeList

		return &frameID, true
	}

	return (*b.replacer).Victim(), false
}

//NewBufferPoolManager returns a empty buffer pool manager
func NewBufferPoolManager(DiskManager disk.DiskManager, clockReplacer *ClockReplacer) *BufferPoolManager {
	freeList := make([]FrameID, 0)
	pages := [MaxPoolSize]*page.Page{}
	for i := 0; i < MaxPoolSize; i++ {
		freeList = append(freeList, FrameID(i))
		pages[FrameID(i)] = nil
	}
	return &BufferPoolManager{DiskManager, pages, clockReplacer, freeList, make(map[page.PageID]FrameID)}
}