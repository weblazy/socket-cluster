package unsafehash

import (
	"sort"

	"github.com/spf13/cast"
	"github.com/weblazy/easy/sortx"
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
	segmentList *sortx.SortList
}

func NewSegmentHash(segmentList ...*Segment) *SegmentHash {
	list := sortx.NewSortList(sortx.ASC)
	for k1 := range segmentList {
		list.List = append(list.List, sortx.Sort{
			Obj:  segmentList[k1],
			Sort: cast.ToFloat64(segmentList[k1].Position),
		})
	}
	sort.Sort(list)
	return &SegmentHash{
		segmentList: list,
	}
}

func (c *SegmentHash) Append(segmentList ...*Segment) {
	for k1 := range segmentList {
		c.segmentList.List = append(c.segmentList.List, sortx.Sort{
			Obj:  segmentList[k1],
			Sort: cast.ToFloat64(segmentList[k1].Position),
		})
	}
	sort.Sort(c.segmentList)
}

func (c *SegmentHash) Get(key int64) interface{} {
	if len(c.segmentList.List) == 0 || cast.ToFloat64(key) > c.segmentList.List[len(c.segmentList.List)-1].Sort {
		// out of range
		return nil
	}
	index := c.search(key)
	i := key % int64(len(c.segmentList.List[index].Obj.(*Segment).List))
	return c.segmentList.List[index].Obj.(*Segment).List[i]
}

func (c *SegmentHash) search(key int64) int {
	n := len(c.segmentList.List)
	i := sort.Search(n, func(i int) bool { return c.segmentList.List[i].Sort >= cast.ToFloat64(key) })
	if i < n {
		return i
	} else {
		return 0
	}
}
