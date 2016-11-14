package main

import "strconv"
import "strings"
import "github.com/gedex/inflector"

func (ic *Compiler) expression() string {
	var token = ic.Scan(0)
	
	for token == "\n" {
		token = ic.Scan(0)
	}
	
	switch token {
		case "true":
			ic.ExpressionType = Number
			return "1"
		case "false":
			ic.ExpressionType = Number
			return "0"
		case "error":
			ic.ExpressionType = Number
			return "ERROR"
	}
	
	if token == "{" {
		var t = ic.Scan(0)
		if t == "}" {
			ic.ExpressionType = Something
		} else {
			ic.ExpressionType = ic.DefinedInterfaces[t].GetType()
			ic.Scan('}')
		}
		var tmp = ic.Tmp("something")
		ic.Assembly("ARRAY ", tmp)
		ic.Assembly("PUT 0")
		return tmp
	}
	
	//Text.
	if token[0] == '"' {
		ic.ExpressionType = Text
		return ic.ParseString(token)
	}
	
	//Letters.
	if token[0] == "'"[0] {
		if s, err := strconv.Unquote(token); err == nil {
			ic.ExpressionType = Letter
			return strconv.Itoa(int([]byte(s)[0]))
		} else {
			ic.RaiseError(err)
		}
	}
	
	//Hexadecimal.
	if len(token) > 2 && token[0] == '0' && token[1] == 'x' { 
		ic.ExpressionType = Number
		return token
	}
	
	//Arrays.
	if token == "[" {
		return ic.ScanArray()
	}
	
	//Pipes.
	if token == "|" {
		var name = "open"
		if ic.Peek() != "|" {
			var arg = ic.ScanExpression()
			name += "_m_"+ic.ExpressionType.Name
			if f, ok := ic.DefinedFunctions[name]; ok {
				var tmp = ic.Tmp("open")
				ic.Assembly(ic.ExpressionType.Push, " ", arg)
				ic.Assembly(ic.RunFunction(name))
				ic.Assembly(f.Returns[0].Pop, " ", tmp)
				ic.ExpressionType = f.Returns[0]
				ic.Scan('|')
				return tmp
			} else {
				ic.RaiseError("Cannot create a pipe out of a ", ic.ExpressionType.Name)
			}
		} else {
			ic.Scan('|')
			var tmp = ic.Tmp("pipe")
			ic.Assembly("PIPE ", tmp)
			ic.ExpressionType = Pipe
			return tmp
			//ic.RaiseError("Blank pipe!")
		}
	}
	
	//Subexpessions.
	if token == "(" {
		defer func() {
			ic.Scan(')')
		}()
		return ic.ScanExpression()
	}
	
	//Is it a literal number? Then just return it.
	if _, err := strconv.Atoi(token); err == nil{
		ic.ExpressionType = Number
		return token
	}
	
	//Minus.
	if token == "-" {
		ic.NextToken = token
		ic.ExpressionType = Number
		return ic.Shunt("0")
	}
	
	if t := ic.GetVariable(token); t != Undefined {
		ic.ExpressionType = t
		ic.SetVariable(token+"_use", Used)
		if t.User {
			return ic.Shunt(token)
		}
		return token
	}
	
	if t, ok := ic.DefinedTypes[token]; ok {
		ic.ExpressionType = t
		
		if ic.Peek() == "(" || ic.NextToken == "(" {
			ic.Scan('(')
			ic.Scan(')')
				
			var array = ic.Tmp("user")
			ic.Assembly("ARRAY ", array)
			for range ic.DefinedTypes[token].Detail.Elements {
				ic.Assembly("PUT 0")
			}
			return array
		} else if ic.Peek() == "{" {
			ic.NextToken = token
			variable := ic.ScanConstructor()
				//TODO better gc protection.
			ic.SetVariable(variable, t)
			ic.SetVariable(variable+"_use", Used)
			return variable
			
		} else if ic.GetFlag(InMethod) && ic.LastDefinedType.Super == token {
			ic.ExpressionType = ic.DefinedTypes[ic.LastDefinedType.Super]
			return ic.LastDefinedType.Name
			
		} else {
			ic.RaiseError()
		}
		
		
	}
	
	if token == "new" {
		var sort = ic.expression()
		if _, ok := ic.DefinedFunctions["new_m_"+ic.ExpressionType.Name]; !ok {
			ic.RaiseError("no new method found for ", ic.ExpressionType.Name)
		}
		var r = ic.Tmp("new")
		ic.Assembly(ic.ExpressionType.Push, " ", sort)
		ic.Assembly("RUN new_m_"+ic.ExpressionType.Name)
		ic.Assembly("GRAB ", r)
		return r
	}
	
	if _, ok := ic.DefinedFunctions[token]; ok {
		if ic.Peek() == "@" {
			ic.Scan('@')
			var variable = ic.expression()
			ic.Assembly("%v %v", ic.ExpressionType.Push, variable)
			var name = ic.ExpressionType.Name
			ic.ExpressionType = InFunction
			return token+"_m_"+name
		}
	
		if ic.Peek() != "(" {
			ic.ExpressionType = Func
			var id = ic.Tmp("func")
			ic.Assembly("SCOPE ", token)
			ic.Assembly("TAKE ", id)
		
			return id
		} else {
			ic.ExpressionType = InFunction
			return token
		}
	}
	
	if ic.Translation && !ic.Translated {
		ic.Translated = true
		defer func() {
			ic.Translated = false
		}()
		var err error
		ic.NextToken, err = getTranslation(ic.Language, "en", token)
		if strings.Contains(ic.NextToken, "\n") {
			ic.NextToken = strings.Split(ic.NextToken, "\n")[0]
		}
		println(ic.NextToken, ic.Language)
		if err != nil {
			ic.RaiseError(err)
		}
		return ic.expression() 
	}
	
	token = inflector.Singularize(token)
	if t, ok := ic.DefinedTypes[token]; ok {
		ic.Scan('(')
		ic.Scan(')')
		return ic.NewListOf(t)
	}
	if t, ok := ic.DefinedInterfaces[token]; ok {
		ic.Scan('(')
		ic.Scan(')')
		return ic.NewListOf(t.GetType())
	}
	
	ic.RaiseError()
	return ""
}

func (ic *Compiler) ScanExpression() string {
	return ic.Shunt(ic.expression())
}
