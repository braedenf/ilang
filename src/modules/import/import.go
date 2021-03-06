package imp

import "github.com/qlova/ilang/src"
import "os"
import "text/scanner"
import "os/user"

func init() {
	ilang.RegisterToken([]string{"import"}, ScanImport)
}

//Messy.
func ScanImport(ic *ilang.Compiler) {
	pkg := ic.Scan(ilang.Name)
	ic.Scan('\n')
	
	var filename = ""
	
	retry:
	file, err := os.Open(pkg+".i")
	if err != nil {
		if file, err = os.Open(pkg+"/"+pkg+".i"); err != nil {
			
			//Search through parent folders?
			dir, _ := os.Getwd()
			if ic.FileDepth > 0 {
				ic.FileDepth--
				os.Chdir(ic.Dirs[len(ic.Dirs)-1])
				ic.Dirs = ic.Dirs[:len(ic.Dirs)-1]
				goto retry
			}
			
			//Search in ~/.ilang.
			if ic.FileDepth == 0 {
				ic.FileDepth--
				usr, err := user.Current()
				if err == nil {
					os.Chdir(usr.HomeDir+"/.ilang/imports/")
					ic.Dirs = append(ic.Dirs, usr.HomeDir+"/.ilang/imports/")
					goto retry
				}
			}
			
			ic.RaiseError("Cannot import "+pkg+", does not exist!", dir)
		} else {
			 filename = pkg+"/"+pkg+".i"
			 
			 dir, _ := os.Getwd()
			 
			ic.Dirs = append(ic.Dirs, dir)
			 
			os.Chdir("./"+pkg)
			ic.FileDepth++
		}
	} else {
		filename = pkg+".i"
	}
	ic.Scanners = append(ic.Scanners, ic.Scanner)
	
	ic.Scanner = &scanner.Scanner{}
	ic.Scanner.Init(file)
	ic.Scanner.Position.Filename = filename
	ic.Scanner.Whitespace= 1<<'\t' | 1<<'\r' | 1<<' '
}
