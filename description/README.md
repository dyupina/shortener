# Optimizations

- A memory consumption profile was initiated. For this purpose, the application was launched and in another terminal, a script was started to load the application.
```
bash test.sh &
curl -sK -v http://localhost:8080/debug/pprof/heap?seconds=30 > profiles/base.pprof
go tool pprof -http=":8082" -seconds=30 profiles/base.pprof
```
- Then the command `top` was entered in the pprof console.
```
(pprof) top
Showing nodes accounting for 16.12kB, 1.55% of 1040.21kB total
Showing top 10 nodes out of 25
      flat  flat%   sum%        cum   cum%
 528.17kB 50.77% 50.77%   528.17kB 50.77%  shortener/internal/user.(*user).AddURLs
-512.05kB 49.23%  1.55%  -512.05kB 49.23%  golang.org/x/text/internal/language.parseVariants
 ...
```
- It's evident that `AddURLs()` consumes the most memory.
- The result of the list `AddURLs()` command highlighted the most problematic area in the code, which is the handling of the map. As an optimization, memory was pre-allocated for the map values.
- After the optimization and running the command `go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof`, the results with negative values are visible, indicating a reduction in resource consumption.
