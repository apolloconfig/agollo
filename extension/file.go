/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package extension

import (
	"container/list"
	"github.com/apolloconfig/agollo/v4/env/file"
)

type HandlerWithPriority struct {
	Handler  file.FileHandler
	priority int
}

var handlers = list.New()

// AddFileHandler 添加一个 FileHandler 实现，并设定其优先级
func AddFileHandler(handler file.FileHandler, priority int) {
	newHandler := HandlerWithPriority{handler, priority}

	// 在链表中找到合适的位置插入
	for e := handlers.Front(); e != nil; e = e.Next() {
		if e.Value.(HandlerWithPriority).priority < priority {
			handlers.InsertBefore(newHandler, e)
			return
		}
	}
	// 如果没有找到合适的位置，追加到链表末尾
	handlers.PushBack(newHandler)
}

// SetFileHandler 清空当前的 handlers 列表，并设置一个新的 FileHandler
func SetFileHandler(handler file.FileHandler) {
	// 清空当前的 handlers 列表
	handlers = list.New()
	// 添加新的 handler
	newHandler := HandlerWithPriority{
		Handler:  handler,
		priority: 10, // 默认优先级
	}
	// 将新的处理器直接添加到 handlers 列表中
	handlers.PushBack(newHandler)
}

// GetFileHandlers 返回按优先级排好序的所有的 FileHandler（priority 值越大，优先级越高）
func GetFileHandlers() *list.List {
	return handlers
}
