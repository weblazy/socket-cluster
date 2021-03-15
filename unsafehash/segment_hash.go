package unsafehash

import (
	"sort"

	esaySort "github.com/weblazy/easy/utils/sort"
)

type Segment struct {
	Position int64
	List     []interface{} // client list
}

func NewSegment(position int64, list []interface{}) *Segment {
	return &Segment{
		Position: position,
		List:     list,
	}
}

type SegmentHash struct {
	segmentList []esaySort.Sort
}

func NewSegmentHash(segmentList ...*Segment) *SegmentHash {
	list := make([]esaySort.Sort, len(segmentList))
	for k1 := range segmentList {
		list = append(list, esaySort.Sort{
			Obj:  segmentList[k1],
			Sort: segmentList[k1].Position,
		})
	}
	sort.Sort(esaySort.SortList(list))
	return &SegmentHash{
		segmentList: list,
	}
}

func (c *SegmentHash) Append(segmentList ...*Segment) {
	for k1 := range segmentList {
		c.segmentList = append(c.segmentList, esaySort.Sort{
			Obj:  segmentList[k1],
			Sort: segmentList[k1].Position,
		})
	}
	sort.Sort(esaySort.SortList(c.segmentList))
}

func (c *SegmentHash) Get(key int64) interface{} {
	if len(c.segmentList) == 0 || key > c.segmentList[len(c.segmentList)-1].Sort {
		// out of range
		return nil
	}
	index := c.search(key)
	i := key % int64(len(c.segmentList[index].Obj.(*Segment).List))
	return c.segmentList[index].Obj.(*Segment).List[i]
}

func (c *SegmentHash) search(key int64) int {
	n := len(c.segmentList)
	i := sort.Search(n, func(i int) bool { return c.segmentList[i].Sort >= key })
	if i < n {
		return i
	} else {
		return 0
	}
}
