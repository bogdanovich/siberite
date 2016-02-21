/*
 * Go memcache client package
 *
 * Author: Krzysztof Kliś <krzysztof.klis@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version with the following modification:
 *
 * As a special exception, the copyright holders of this library give you
 * permission to link this library with independent modules to produce an
 * executable, regardless of the license terms of these independent modules,
 * and to copy and distribute the resulting executable under terms of your choice,
 * provided that you also meet, for each linked independent module, the terms
 * and conditions of the license of that module. An independent module is a
 * module which is not derived from or based on this library. If you modify this
 * library, you may extend this exception to your version of the library, but
 * you are not obligated to do so. If you do not wish to do so, delete this
 * exception statement from your version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package gomemcache

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Memcache struct {
	conn net.Conn
}

type Result struct {
	Value []uint8
	Flags int
}

var (
	ConnectionError = errors.New("memcache: not connected")
	ReadError       = errors.New("memcache: read error")
	DeleteError     = errors.New("memcache: delete error")
	FlushAllError   = errors.New("memcache: flush_all error")
	NotFoundError   = errors.New("memcache: not found")
)

func Connect(host string, port int) (*Memcache, error) {
	var network, addr string
	if port == 0 {
		network = "unix"
		addr = host
	} else {
		network = "tcp"
		addr = host + ":" + strconv.Itoa(port)
	}
	return Dial(network, addr)
}

func Dial(network, addr string) (memc *Memcache, err error) {
	memc = new(Memcache)
	conn, err := net.Dial(network, addr)
	if err != nil {
		return
	}
	memc.conn = conn
	return
}

func (memc *Memcache) Close() (err error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	return memc.conn.Close()
}

func (memc *Memcache) FlushAll() (err error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	cmd := "flush_all\r\n"
	_, err1 := memc.conn.Write([]uint8(cmd))
	if err1 != nil {
		err = err1
		return err
	}
	reader := bufio.NewReader(memc.conn)
	line, err1 := reader.ReadString('\n')
	if err1 != nil {
		err = err1
		return err
	}
	if line != "OK\r\n" {
		return FlushAllError
	}
	return nil
}

func (memc *Memcache) Get(key string) (value []byte, flags int, err error) {
	if memc == nil || memc.conn == nil {
		err = ConnectionError
		return
	}
	cmd := "get " + key + "\r\n"
	_, err = memc.conn.Write([]uint8(cmd))
	if err != nil {
		return
	}
	reader := bufio.NewReader(memc.conn)
	return memc.readValue(reader, key)
}

func (memc *Memcache) GetMulti(keys ...string) (results map[string]Result, err error) {
	results = map[string]Result{}
	for _, key := range keys {
		value, flags, err1 := memc.Get(key)
		if err1 == nil {
			results[key] = Result{Value: value, Flags: flags}
		} else if err1 != NotFoundError {
			err = err1
			return
		}
	}
	return
}

func (memc *Memcache) readValue(reader *bufio.Reader, key string) (value []byte, flags int, err error) {
	line, err1 := reader.ReadString('\n')
	if err1 != nil {
		err = err1
		return
	}
	a := strings.Split(strings.TrimSpace(line), " ")
	if len(a) != 4 || a[0] != "VALUE" {
		if line == "END\r\n" {
			err = NotFoundError
		} else {
			err = ReadError
		}
		return
	}
	flags, _ = strconv.Atoi(a[2])
	l, _ := strconv.Atoi(a[3])
	value = make([]byte, l)
	n := 0
	for {
		i, err1 := reader.Read(value[n:])
		if i == 0 && err == io.EOF {
			break
		}
		if err1 != nil {
			err = err1
			return
		}
		n += i
		if n >= l {
			break
		}
	}
	if n != l {
		err = ReadError
		return
	}
	line, err = reader.ReadString('\n')
	if err != nil {
		return
	}
	if line != "\r\n" {
		err = ReadError
		return
	}
	return
}

func (memc *Memcache) store(cmd string, key string, value []byte, flags int, exptime int64) (err error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	l := len(value)
	s := cmd + " " + key + " " + strconv.Itoa(flags) + " " + strconv.FormatInt(exptime, 10) + " " + strconv.Itoa(l) + "\r\n"
	writer := bufio.NewWriter(memc.conn)
	_, err = writer.WriteString(s)
	if err != nil {
		return err
	}
	_, err = writer.Write(value)
	if err != nil {
		return err
	}
	_, err = writer.WriteString("\r\n")
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	reader := bufio.NewReader(memc.conn)
	line, err1 := reader.ReadString('\n')
	if err1 != nil {
		err = err1
		return err
	}
	if line != "STORED\r\n" {
		WriteError := errors.New("memcache: " + strings.TrimSpace(line))
		return WriteError
	}
	return nil
}

func (memc *Memcache) Set(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("set", key, value, flags, exptime)
}

func (memc *Memcache) Add(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("add", key, value, flags, exptime)
}

func (memc *Memcache) Replace(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("replace", key, value, flags, exptime)
}

func (memc *Memcache) Append(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("append", key, value, flags, exptime)
}

func (memc *Memcache) Prepend(key string, value []byte, flags int, exptime int64) (err error) {
	return memc.store("prepend", key, value, flags, exptime)
}

func (memc *Memcache) Delete(key string) (err error) {
	if memc == nil || memc.conn == nil {
		return ConnectionError
	}
	cmd := "delete " + key + "\r\n"
	_, err1 := memc.conn.Write([]uint8(cmd))
	if err1 != nil {
		err = err1
		return err
	}
	reader := bufio.NewReader(memc.conn)
	line, err1 := reader.ReadString('\n')
	if err1 != nil {
		err = err1
		return err
	}
	if line != "DELETED\r\n" {
		return DeleteError
	}
	return nil
}

func (memc *Memcache) incdec(cmd string, key string, value uint64) (i uint64, err error) {
	if memc == nil || memc.conn == nil {
		err = ConnectionError
		return
	}
	s := cmd + " " + key + " " + strconv.FormatUint(value, 10) + "\r\n"
	_, err = memc.conn.Write([]uint8(s))
	if err != nil {
		return
	}
	reader := bufio.NewReader(memc.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	if line == "NOT_FOUND\r\n" {
		err = NotFoundError
		return
	}
	i, err = strconv.ParseUint(strings.TrimSpace(line), 10, 64)
	return
}

func (memc *Memcache) Incr(key string, value uint64) (i uint64, err error) {
	i, err = memc.incdec("incr", key, value)
	return
}

func (memc *Memcache) Decr(key string, value uint64) (i uint64, err error) {
	i, err = memc.incdec("decr", key, value)
	return
}

func (memc *Memcache) SetReadTimeout(nsec int64) (err error) {
	return memc.conn.SetReadDeadline(time.Now().Add(time.Duration(nsec)))
}

func (memc *Memcache) SetWriteTimeout(nsec int64) (err error) {
	return memc.conn.SetWriteDeadline(time.Now().Add(time.Duration(nsec)))
}
