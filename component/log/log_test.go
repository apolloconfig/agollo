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

package log

import "testing"

func TestDebugf(t *testing.T) {
	Logger.Debugf("")
}

func TestInfof(t *testing.T) {
	Logger.Infof("")
}

func TestErrorf(t *testing.T) {
	Logger.Errorf("")
}

func TestWarnf(t *testing.T) {
	Logger.Warnf("")
}

func TestDebug(t *testing.T) {
	Logger.Debug("")
}

func TestInfo(t *testing.T) {
	Logger.Info("")
}

func TestError(t *testing.T) {
	Logger.Error("")
}

func TestWarn(t *testing.T) {
	Logger.Warn("")
}

func TestInitLogger(t *testing.T) {
	InitLogger(Logger)
}

func TestCommonDebugf(t *testing.T) {
	Debugf("")
}

func TestCommonInfof(t *testing.T) {
	Infof("")
}

func TestCommonErrorf(t *testing.T) {
	Errorf("")
}

func TestCommonWarnf(t *testing.T) {
	Warnf("")
}

func TestCommonDebug(t *testing.T) {
	Debug("")
}

func TestCommonInfo(t *testing.T) {
	Info("")
}

func TestCommonError(t *testing.T) {
	Error("")
}

func TestCommonWarn(t *testing.T) {
	Warn("")
}
