package utils_test

import (
	"github.com/OffchainLabs/prysm/v6/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

func Test_RWMap(t *testing.T) {
	testRWMap(t, utils.NewRWMap[int, string]())
	testRWMap(t, utils.NewRWMapFromStdMap(map[int]string{}))
}

func testRWMap(t *testing.T, rwMap *utils.RWMap[int, string]) {
	require.Equal(t, map[int]string{}, rwMap.LoadAll())
	var wg sync.WaitGroup

	data := make(map[int]string)
	for i := 0; i < 100; i++ {
		data[i] = uuid.NewString()
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			rwMap.Store(i, data[i])

			lVal, ok := rwMap.Load(i)
			require.True(t, ok)
			require.Equal(t, lVal, data[i])

			rwMap.Delete(i)
			lVal, ok = rwMap.Load(i)
			require.False(t, ok)
			require.Equal(t, "", lVal)

			rwMap.Store(i, data[i])
		}(i)
	}
	wg.Wait()

	require.Equal(t, data, rwMap.LoadAll())
	rwMap.DeleteAll()
	require.Equal(t, map[int]string{}, rwMap.LoadAll())

	t.Run("should replace map", func(t *testing.T) {
		// given
		rwMap.Store(1, "one")
		newMap := map[int]string{
			2: "two",
		}

		// when
		rwMap.Replace(newMap)

		// then
		_, ok := rwMap.Load(1)
		require.False(t, ok, "Expected 'one' to be deleted after replace")

		val, ok := rwMap.Load(2)
		require.True(t, ok && val == "two", "Expected to load (2, true) after replace, got (%s, %v)", val, ok)

		t.Run("concurrent replace", func(t *testing.T) {
			// given
			var wgR sync.WaitGroup
			wgR.Add(100)

			// when
			for i := 0; i < 100; i++ {
				go func(j int) {
					defer wgR.Done()
					rwMap.Replace(map[int]string{
						j: uuid.NewString(),
					})
				}(i)
			}

			// then
			wg.Wait()
		})
	})
}
