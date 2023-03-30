/*
 *
 *     @file: Log_test.go
 *     @author: Equationzhao
 *     @email: equationzhao@foxmail.com
 *     @time: 2023/3/30 下午11:29
 *     @last modified: 2023/3/30 下午3:37
 *
 *
 *
 */

package Log

import (
	"testing"
)

func TestInfo(t *testing.T) {
	Infof("hello %s", "world")

	Info("hello", "toWhom", "world")

	Info("hello", "toWhom", "world", "age", 18)

	Info("hello", String("toWhom", "world"), Int("age", 18), Bool("isMale", true))
}
