package buffer

import (
	"crypto/rand"
	"testing"

	"github.com/brunocalza/go-bustub/storage/disk"
	"github.com/brunocalza/go-bustub/storage/page"
	"github.com/brunocalza/go-bustub/testingutils"
)

func TestBinaryData(t *testing.T) {
	poolSize := uint32(10)

	diskManager := disk.NewDiskManagerTest()
	defer diskManager.ShutDown()
	bpm := NewBufferPoolManager(poolSize, diskManager)

	page0 := bpm.NewPage()

	// Scenario: The buffer pool is empty. We should be able to create a new page.
	testingutils.Equals(t, page.PageID(0), page0.ID())

	// Generate random binary data
	randomBinaryData := make([]byte, page.PageSize)
	rand.Read(randomBinaryData)

	// Insert terminal characters both in the middle and at end
	randomBinaryData[page.PageSize/2] = '0'
	randomBinaryData[page.PageSize-1] = '0'

	var fixedRandomBinaryData [page.PageSize]byte
	copy(fixedRandomBinaryData[:], randomBinaryData[:page.PageSize])

	// Scenario: Once we have a page, we should be able to read and write content.
	page0.CopyToData(randomBinaryData)
	testingutils.Equals(t, fixedRandomBinaryData, *page0.Data())

	// Scenario: We should be able to create new pages until we fill up the buffer pool.
	for i := uint32(1); i < poolSize; i++ {
		p := bpm.NewPage()
		testingutils.Equals(t, page.PageID(i), p.ID())
	}

	// Scenario: Once the buffer pool is full, we should not be able to create any new pages.
	for i := poolSize; i < poolSize*2; i++ {
		testingutils.Equals(t, (*page.Page)(nil), bpm.NewPage())
	}

	// Scenario: After unpinning pages {0, 1, 2, 3, 4} and pinning another 4 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		testingutils.Ok(t, bpm.UnpinPage(page.PageID(i), true))
		bpm.FlushPage(page.PageID(i))
	}
	for i := 0; i < 4; i++ {
		p := bpm.NewPage()
		bpm.UnpinPage(p.ID(), false)
	}

	// Scenario: We should be able to fetch the data we wrote a while ago.
	page0 = bpm.FetchPage(page.PageID(0))
	testingutils.Equals(t, fixedRandomBinaryData, *page0.Data())
	testingutils.Ok(t, bpm.UnpinPage(page.PageID(0), true))
}

func TestSample(t *testing.T) {
	poolSize := uint32(10)

	diskManager := disk.NewDiskManagerTest()
	defer diskManager.ShutDown()
	bpm := NewBufferPoolManager(poolSize, diskManager)

	page0 := bpm.NewPage()

	// Scenario: The buffer pool is empty. We should be able to create a new page.
	testingutils.Equals(t, page.PageID(0), page0.ID())

	// Scenario: Once we have a page, we should be able to read and write content.
	page0.CopyToData([]byte("Hello"))
	testingutils.Equals(t, [page.PageSize]byte{'H', 'e', 'l', 'l', 'o'}, *page0.Data())

	// Scenario: We should be able to create new pages until we fill up the buffer pool.
	for i := uint32(1); i < poolSize; i++ {
		p := bpm.NewPage()
		testingutils.Equals(t, page.PageID(i), p.ID())
	}

	// Scenario: Once the buffer pool is full, we should not be able to create any new pages.
	for i := poolSize; i < poolSize*2; i++ {
		testingutils.Equals(t, (*page.Page)(nil), bpm.NewPage())
	}

	// Scenario: After unpinning pages {0, 1, 2, 3, 4} and pinning another 4 new pages,
	// there would still be one cache frame left for reading page 0.
	for i := 0; i < 5; i++ {
		testingutils.Ok(t, bpm.UnpinPage(page.PageID(i), true))
		bpm.FlushPage(page.PageID(i))
	}
	for i := 0; i < 4; i++ {
		bpm.NewPage()
	}
	// Scenario: We should be able to fetch the data we wrote a while ago.
	page0 = bpm.FetchPage(page.PageID(0))
	testingutils.Equals(t, [page.PageSize]byte{'H', 'e', 'l', 'l', 'o'}, *page0.Data())

	// Scenario: If we unpin page 0 and then make a new page, all the buffer pages should
	// now be pinned. Fetching page 0 should fail.
	testingutils.Ok(t, bpm.UnpinPage(page.PageID(0), true))

	testingutils.Equals(t, page.PageID(14), bpm.NewPage().ID())
	testingutils.Equals(t, (*page.Page)(nil), bpm.NewPage())
	testingutils.Equals(t, (*page.Page)(nil), bpm.FetchPage(page.PageID(0)))
}