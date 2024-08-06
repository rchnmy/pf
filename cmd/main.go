package main

import (
    "os"
    "fmt"
    "math"
    "io/fs"
    "slices"
    "strconv"

    "golang.org/x/sys/unix"
)

func main() {
    if len(os.Args) == 1 {
        fmt.Println("Usage: pf PID")
        return
    }
    arg := os.Args[1]
    pid, err := strconv.ParseUint(arg, 10, 32)
    if err != nil {
        fmt.Println(err.(*strconv.NumError).Err)
        return
    }
    if _, err := unix.PidfdOpen(int(pid), 0); err != nil {
       fmt.Println(err)
       return
    }

    root := "/proc/" + arg + "/fd"
    if !fs.ValidPath(root[1:]) {
        fmt.Printf("%s not found\n", root)
        return
    }
    fd, err := fs.ReadDir(os.DirFS(root), ".")
    if err != nil {
        fmt.Println(err.(*fs.PathError).Err)
        return
    }
    if len(fd) < 4 {
        fmt.Println("no regular files")
        return
    }

    files := make(map[int64]string, 0)
    for i := 3; i < len(fd); i++ {
        buf := make([]byte, 255)
        if _, err := unix.Readlink(root + "/" + fd[i].Name(), buf); err != nil {
            fmt.Println(err)
            return
        }
        file := unix.ByteSliceToString(buf)
        stat := &unix.Stat_t{}
        if err := unix.Stat(file, stat); err == nil && isRegular(stat.Mode) {
            files[stat.Size] = file
        }
    }
    if len(files) == 0 {
        fmt.Println("no regular files")
        return
    }

    sizes := make([]int64, 0, len(files))
    for size := range files {
        sizes = append(sizes, size)
    }
    slices.Sort(sizes); slices.Reverse(sizes)
    for _, size := range sizes {
        fmt.Printf("%-7s %s\n", humanize(size), files[size])
    }
}

func isRegular(mode uint32) bool {
    return (mode & unix.S_IFMT) == unix.S_IFREG
}

func humanize(size int64) string {
    if size < 10 {
        return fmt.Sprintf("%dB", size)
    }
    e := math.Floor(math.Log(float64(size)) / math.Log(1000))
    n := float64(size) / math.Pow(1000, e)
    u := []string{"B", "K", "M", "G", "T", "P", "E"}
    return fmt.Sprintf("%.1f%s", n, u[int(e)])
}

