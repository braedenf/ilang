package main

import "text/scanner"
import (
	"os"
	"fmt"
	"strings"
	"strconv"
	"io"
	"flag"
)

//These are the 4 types in I.
const (
	FUNCTION = iota
	STRING
	NUMBER
	FILE
)

//This holds the definition of a function.
type Function struct {
	Exists bool
	Args []int
	Returns []int
	
	//Is this a local?
	Local bool
	
	Variadic bool
}

var variables = make( map[string]bool)
var functions = make( map[string]Function)
var unique int

func expression(s *scanner.Scanner, output io.Writer, param ...bool) string {
	
	var shunting bool = len(param) <= 0 || param[0]

	//Turn string literals into numeric strings.
	//For example string arguments to a function
	//eg. output("A")
	// ->
	// STRING i+tmp+id
	// PUSH 'A' i+tmp+id
	// PUSHSTRING i+tmp+id
	// RUN output.
	if len(s.TokenText()) > 0 && s.TokenText()[0] == '"' {
				
		unique++
		var newarg string = "STRING i+tmp+"+fmt.Sprint(unique)+"\n"
		var j int
		var arg = s.TokenText()[1:]
		
		stringloop:
		arg = strings.Replace(arg, "\\n", "\n", -1)
		for _, v := range arg {
			if v == '"' {
				goto end
			}
			newarg += "PUSH "+strconv.Itoa(int(v))+" i+tmp+"+fmt.Sprint(unique)+"\n"
		}
		if len(arg) == 0 {
			goto end
		}
		newarg += "PUSH "+strconv.Itoa(int(' '))+" i+tmp+"+fmt.Sprint(unique)+"\n"
		j++
		//println(arg)
		arg = string(s.TokenText()[j])
		goto stringloop
		end:
		//println(newarg)
		output.Write([]byte(newarg))
		if shunting {
			return shunt("i+tmp+"+fmt.Sprint(unique), s, output)
		} else {
			return "i+tmp+"+fmt.Sprint(unique)
		}
	}
	
	if  s.TokenText()[0] == '[' {
		unique++
		var id = unique
		
		output.Write([]byte("STRING i+string+"+fmt.Sprint(id)+"\n"))
		
		for tok := s.Scan(); tok != scanner.EOF; {
		
			if s.TokenText() == "]" {
				break
			}
		
			output.Write([]byte("PUSH "+expression(s, output)+" i+string+"+fmt.Sprint(id)+"\n"))
			
			if s.TokenText() == "]" {
				break
			}
			
			if s.TokenText() != "," {
				fmt.Println(s.Pos(), "Expecting , found ", s.TokenText())
				os.Exit(1)
			}
			s.Scan()
		}
		if shunting {
			return shunt("i+string+"+fmt.Sprint(unique), s, output)
		} else {
			return "i+string+"+fmt.Sprint(unique)
		}
	}
	
	if len(s.TokenText()) == 3 && s.TokenText()[0] == '\'' && s.TokenText()[2] == '\'' {
		defer s.Scan()
		return strconv.Itoa(int(s.TokenText()[1]))
	} else if s.TokenText() == `'\n'` {
		defer s.Scan()
		return strconv.Itoa(int('\n'))
	}


	//Is it a literal number?
	if _, err := strconv.Atoi(s.TokenText()); err == nil {
		if shunting {
			return shunt(s.TokenText(), s, output)
		} else {
			return s.TokenText()
		}
	} else {
	
		var name = s.TokenText()
	
		//Function call.
		if functions[name].Exists  {
		
			if functions[name].Variadic {
				unique++
				var id = unique
				output.Write([]byte("STRING i+variadic+"+fmt.Sprint(unique)+"\n"))
				for tok := s.Scan(); tok != scanner.EOF; {
					
					if s.TokenText() == ")" {
						break
					}
					s.Scan()
				
					output.Write([]byte("PUSH "+expression(s, output)+" i+variadic+"+fmt.Sprint(id)+"\n"))
					
					if s.TokenText() == ")" {
						break
					}
					
					if s.TokenText() != "," {
						fmt.Println(s.Pos(), "Expecting , found ", s.TokenText())
						os.Exit(1)
					}
				}
			
				output.Write([]byte("PUSHSTRING i+variadic+"+fmt.Sprint(id)+"\n"))
				if functions[name].Local {
					output.Write([]byte("EXE "+name+"\n"))
				} else {
					output.Write([]byte("RUN "+name+"\n"))
				}
				if len(functions[name].Returns) > 0 {
					unique++
					switch functions[name].Returns[0] {
						case STRING:
							output.Write([]byte("POPSTRING i+output+"+fmt.Sprint(unique)+"\n"))
						case NUMBER:
							output.Write([]byte("POP i+output+"+fmt.Sprint(unique)+"\n"))
						case FUNCTION:
							output.Write([]byte("POPFUNC i+output+"+fmt.Sprint(unique)+"\n"))
						case FILE:
							output.Write([]byte("POPIT i+output+"+fmt.Sprint(unique)+"\n"))
					}
				}	
				if shunting {
					return shunt("i+output+"+fmt.Sprint(unique), s, output)
				} else {
					return "i+output+"+fmt.Sprint(unique)
				}
			}

			var i int
			for tok := s.Scan(); tok != scanner.EOF; {
				if s.TokenText() == ")" {
					return name
				}
				s.Scan()
				if s.TokenText() == ")" {
					break
				}
				
				if len(functions[name].Args) > i {
					switch functions[name].Args[i] {
						case STRING:
							output.Write([]byte("PUSHSTRING "+expression(s, output)+"\n"))
						case NUMBER:
							output.Write([]byte("PUSH "+expression(s, output)+"\n"))
						case FUNCTION:
							output.Write([]byte("PUSHFUNC "+expression(s, output)+"\n"))
						case FILE:
							output.Write([]byte("PUSHIT "+expression(s, output)+"\n"))
					}
				} 
				
				if s.TokenText() == ")" {
					break
				}
				if s.TokenText() != "," {
					fmt.Println(s.Pos(), "Expecting , found ", s.TokenText())
					os.Exit(1)
				}
				
				
			}		
			if functions[name].Local {
				output.Write([]byte("EXE "+name+"\n"))
			} else {
				output.Write([]byte("RUN "+name+"\n"))
			}
			if len(functions[name].Returns) > 0 {
				unique++
				switch functions[name].Returns[0] {
					case STRING:
						output.Write([]byte("POPSTRING i+output+"+fmt.Sprint(unique)+"\n"))
					case NUMBER:
						output.Write([]byte("POP i+output+"+fmt.Sprint(unique)+"\n"))
					case FUNCTION:
						output.Write([]byte("POPFUNC i+output+"+fmt.Sprint(unique)+"\n"))
					case FILE:
						output.Write([]byte("POPIT i+output+"+fmt.Sprint(unique)+"\n"))
				}
			}		
			var tmp = unique
			if shunting {
				return shunt("i+output+"+fmt.Sprint(unique), s, output)
			}	
			return "i+output+"+fmt.Sprint(tmp)
		}
	
		//Is it a variable?
		if variables[s.TokenText()] {
			if shunting {
				return shunt(s.TokenText(), s, output)
			} else {
				return s.TokenText()
			}
			
		} else {
			
			// a=2; b=4; ab
			if len(s.TokenText()) > 0 && variables[string(rune(s.TokenText()[0]))] {
				if len(s.TokenText()) == 2 {
					if variables[string(rune(s.TokenText()[1]))] {
						unique++
						output.Write([]byte("VAR i+tmp+"+s.TokenText()+fmt.Sprint(unique)+"\n"))
						output.Write([]byte("MUL i+tmp+"+s.TokenText()+fmt.Sprint(unique)+" "+
							string(rune(s.TokenText()[0]))+" "+
							string(rune(s.TokenText()[1]))+"\n"))
						
						if shunting {
							return shunt("i+tmp+"+s.TokenText()+fmt.Sprint(unique), s, output)
						} else {
							return "i+tmp+"+s.TokenText()+fmt.Sprint(unique)
						}
					}
				}
			}
			
		}
	
	}
	if shunting {
		return shunt(s.TokenText(), s, output)
	} else {
		return s.TokenText()
	}
}

func main() {
	flag.Parse()

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		return
	}
	
	output, err := os.Create(flag.Arg(0)[:len(flag.Arg(0))-2]+".u")
	if err != nil {
		return
	}
	
	//Add builtin functions.
	builtin(output)
	
	var s scanner.Scanner
	s.Init(file)
	s.Whitespace= 1<<'\t' | 1<<'\r' | 1<<' '
	
	var tok rune
	for tok != scanner.EOF {
		tok = s.Scan()
		
		switch s.TokenText() {
			case "\n":
				
			
			case "}":
				output.Write([]byte("END\n"))
				
			//Inline universal assembly.
			case ".":
				s.Scan()
				output.Write([]byte(strings.ToUpper(s.TokenText()+" ")))
				for tok = s.Scan(); tok != scanner.EOF; {
					if s.TokenText() == "\n" {
						output.Write([]byte("\n"))
						break
					}
					output.Write([]byte(s.TokenText()))
					s.Scan()
				}
			
			case "return":
				s.Scan()
				if s.TokenText() != "\n" {
					output.Write([]byte("PUSH "+expression(&s, output)+"\n"))
				}
				output.Write([]byte("RETURN\n"))
			
			case "software":
				output.Write([]byte("ROUTINE\n"))
				s.Scan()
				if s.TokenText() != "{" {
					fmt.Println(s.Pos(), "Expecting { found ", s.TokenText())
					return
				}
				s.Scan()
				if s.TokenText() != "\n" {
					fmt.Println(s.Pos(), "Expecting newline found ", s.TokenText())
					return
				}
			
			case "issues":
				output.Write([]byte("IF ERROR\nADD ERROR 0 0\n"))
				s.Scan()
				if s.TokenText() != "{" {
					fmt.Println(s.Pos(), "Expecting { found ", s.TokenText())
					return
				}
				s.Scan()
				if s.TokenText() != "\n" {
					fmt.Println(s.Pos(), "Expecting newline found ", s.TokenText())
					return
				}
				
			//Compiles function declerations.
			case "function":
				var name string
				var function Function
				
				// function name(param1, param2) returns {
				output.Write([]byte("SUBROUTINE "))
				s.Scan()
				output.Write([]byte(s.TokenText()+"\n"))
				name = s.TokenText()
				s.Scan()
				if s.TokenText() != "(" {
					fmt.Println(s.Pos(), "Expecting ( found ", s.TokenText())
					return
				}
				
				//We need to reverse the POP's because of stack pain.
				var toReverse []string
				for tok = s.Scan(); tok != scanner.EOF; {
					var popstring string
					if s.TokenText() == ")" {
						break
					}
					//String arguments.
					if s.TokenText() == "[" {
						//Update our function definition with a string argument.
						function.Args = append(function.Args, STRING)
						
						popstring += "POPSTRING "
						s.Scan()
						if s.TokenText() != "]" {
							fmt.Println(s.Pos(), "Expecting ] found ", s.TokenText())
							return
						}
						s.Scan()
					//Other type of string argument. (Variadic)
					} else if s.TokenText() == "." {
						
						//Update our function definition with a string argument.
						function.Args = append(function.Args, STRING)
						function.Variadic = true
						
						popstring += "POPSTRING "
						s.Scan()
						if s.TokenText() != "." {
							fmt.Println(s.Pos(), "Expecting . found ", s.TokenText())
							return
						}
						s.Scan()
					//Function arguments.
					} else if s.TokenText() == "(" {
						
						//Update our function definition with a string argument.
						function.Args = append(function.Args, FUNCTION)
						
						popstring += "POPFUNC "
						s.Scan()
						if s.TokenText() != ")" {
							fmt.Println(s.Pos(), "Expecting ) found ", s.TokenText())
							return
						}
						s.Scan()
					} else {
						//Update our function definition with a numeric argument.
						function.Args = append(function.Args, NUMBER)
						
						popstring += "POP "
					}
					popstring += s.TokenText()+"\n"
					toReverse = append(toReverse, popstring)
					s.Scan()
					if s.TokenText() == ")" {
						break
					}
					if s.TokenText() != "," {
						fmt.Println(s.Pos(), "Expecting , found ", s.TokenText())
						return
					}
					s.Scan()
				}
				for i := len(toReverse)-1; i>=0; i-- {
					output.Write([]byte(toReverse[i]))
				}
				s.Scan()
				if s.TokenText() != "{" {
					if s.TokenText() != "[" {
						function.Returns = append(function.Returns, NUMBER)
					} else {
						function.Returns = append(function.Returns, STRING)
						s.Scan()
						if s.TokenText() != "]" {
							fmt.Println(s.Pos(), "Expecting ] found ", s.TokenText())
							return
						}
					}
					s.Scan()
					if s.TokenText() != "{" {	
						fmt.Println(s.Pos(), "Expecting { found ", s.TokenText())
						return
					}
				}
				s.Scan()
				if s.TokenText() != "\n" {
					fmt.Println(s.Pos(), "Expecting newline found ", s.TokenText())
					return
				}
			
				function.Exists = true
				functions[name] = function
			default:
			
				var name = s.TokenText()
				if functions[name].Exists {
					var returns = functions[name].Returns
					var f = functions[name]
					f.Returns = nil
					functions[name] = f
						expression(&s, output)
					f.Returns = returns
					functions[name] = f
					continue
				}
				
				s.Scan()
				switch s.TokenText() {
					case "(":
						s.Scan()
						output.Write([]byte("EXE "+name+" \n"))
					case "&":
						s.Scan()
						variables[name] = true
						output.Write([]byte("PUSH "+expression(&s, output)+" "+name+" \n"))
					case "=":
						// a = 
						s.Scan()
						if s.TokenText() == "[" {
							//a = [12,32,92]
							output.Write([]byte("STRING "+name+"\n"))
							
							for tok = s.Scan(); tok != scanner.EOF; {
							
								if s.TokenText() == "]" {
									break
								}
							
								output.Write([]byte("PUSH "+expression(&s, output)+" "+name+"\n"))
								
								if s.TokenText() == "]" {
									break
								}
								
								if s.TokenText() != "," {
									fmt.Println(s.Pos(), "Expecting , found ", s.TokenText())
									return
								}
								s.Scan()
							}
						
						} else if s.TokenText()[0] == '"' {
							//Turn string literals into numeric strings.
							//For example string arguments to a function
							//eg. output("A")
							// ->
							// STRING i+tmp+id
							// PUSH 'A' i+tmp+id
							// PUSHSTRING i+tmp+id
							// RUN output
								var newarg string = "STRING "+name+"\n"
								var j int
								var arg = s.TokenText()[1:]
		
								stringloop:
								arg = strings.Replace(arg, "\\n", "\n", -1)
								for _, v := range arg {
									if v == '"' {
										goto end
									}
									newarg += "PUSH "+strconv.Itoa(int(v))+" "+name+"\n"
								}
								if len(arg) == 0 {
									goto end
								}
								newarg += "PUSH "+strconv.Itoa(int(' '))+" "+name+"\n"
								j++
								//println(arg)
								arg = string(s.TokenText()[j])
								goto stringloop
								end:
								//println(newarg)
								output.Write([]byte(newarg))
								s.Scan()
						
						} else {
							if functions[s.TokenText()].Exists && s.Peek() != '(' {
								
								functions[name] = functions[s.TokenText()]
								f := functions[name] 
								f.Local = true
								functions[name] = f
								output.Write([]byte("FUNC "+name+" "+s.TokenText()+"\n"))
								
							} else {
						
								variables[name] = true
								output.Write([]byte("VAR "+name+" "+expression(&s, output)+"\n"))
							}
						}
					default:
						if len(s.TokenText()) > 0 && s.TokenText()[0] == '.' {
							var index = s.TokenText()[1:]
							s.Scan()
							if s.TokenText() != "=" {
								fmt.Println(s.Pos(), "Expecting = found ", s.TokenText())
								return
							}
							s.Scan()
							output.Write([]byte("SET "+name+" "+index+" "+expression(&s, output)+"\n"))
							
						} else {
					
					
							if name == "" {
								return	
							}
							fmt.Println(s.Pos(), "Unexpected ", name)
							return
						}
				}
				
		}
	}
}