package proj

import (
	"hash/crc64"
)

func Memoizer(memHandle MnistHandle, classHandle MnistHandle, cacheHandle CacheHandle) {
	defer close(memHandle.RespQ)

	var cachegone bool = false
	var classgone bool = false
	table := crc64.MakeTable(crc64.ECMA)

	for req := range memHandle.ReqQ {
		key := crc64.Checksum(req.Val, table)
		var cresp CacheResp
		var okc bool = true

		if cachegone == false {
			creq := CacheReq{false, key, 0, req.Id}
			cacheHandle.ReqQ <- creq
			cresp, okc = <-cacheHandle.RespQ
			if okc == false {
				cachegone = true
			}
		}

		if cachegone == true || cresp.Exists == false || cresp.Id != req.Id {
			classHandle.ReqQ <- req
			resp, ok := <-classHandle.RespQ
			if ok == false {
				var cause MemErrCause = MemErr_serCrash
				shit := CreateMemErr(cause, "Classifier crashed", nil)
				ans := MnistResp{resp.Val, req.Id, shit}
				memHandle.RespQ <- ans
				classgone = true
			} else if resp.Err != nil {
				var cause MemErrCause = MemErr_serErr
				shit := CreateMemErr(cause, "Classifier error", resp.Err)
				ans := MnistResp{resp.Val, req.Id, shit}
				memHandle.RespQ <- ans
			} else if req.Id != resp.Id {
				var cause MemErrCause = MemErr_serCorrupt
				shit := CreateMemErr(cause, "Bad messageID from classifier", nil)
				ans := MnistResp{resp.Val, req.Id, shit}
				memHandle.RespQ <- ans
			} else {
				if cachegone == false {
					creqw := CacheReq{true, key, resp.Val, resp.Id}
					cacheHandle.ReqQ <- creqw
				}
				memHandle.RespQ <- resp
			}
		} else {
			ans := MnistResp{cresp.Val, cresp.Id, nil}
			memHandle.RespQ <- ans
		}
	}
	if classgone == false {
		close(classHandle.RespQ)
	}
	if cachegone == false {
		close(cacheHandle.RespQ)
	}
}
