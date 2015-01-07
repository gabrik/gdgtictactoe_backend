package hello



/**
 * Richieste tipo
 *
 * http://localhost:8080/appSave?status={"Score":-10,"Email":"turi@aceto.it"}
 * http://localhost:8080/appNewUser?status={"Name":"pippo","Email":"turi@aceto.it"}
 * http://localhost:8080/appHistory?status={"Name":"pippo","Email":"turi@olio.it"}
 * http://localhost:8080/appLeaderboard
 * http://localhost:8080/appConcurrent?status{"Status":"O-O-XX---","Tableau":[[1,0,1],[0,-1,-1],[0,0,0]]}
 *
 *
 */

import (
 "fmt"
 "net/http"

        //"strconv"
 "encoding/json"
 "time"


 "appengine"
 "appengine/datastore"
)

type User struct{

	Name string
	Email string

}


type Result struct{
    Email string
    Score int
}

type Leader struct{
    Name string
    Score int
}


type Game struct{
    Status string
    Tableau [][]int
}


type Score struct{
    score int
    mUser User
}






type ConcGame func(status string) string

func playGame(status string) ConcGame{

	return func(query string) string{
		var g Game
		err := json.Unmarshal([]byte(status), &g)
		if err != nil {
			e,_:=json.Marshal(err)
			return string(e)
		}

		g.play()
		g.makeTableau()
		b, _ := json.Marshal(g)
		return string(b)
	}
	
}

func FirstGame(status string, replicas ...ConcGame) string {
    c := make(chan string)
    playReplica := func(i int) { c <- replicas[i](status) }
    for i := range replicas {
        go playReplica(i)
    }
    return <-c
}



type Load func(context appengine.Context) string


func doLoad(context appengine.Context) Load{
   return func(context appengine.Context) string{
        return makeLeaderBoard(context)
   }
   
}

func makeLeaderBoard (context appengine.Context) string {

    users := make([]User, 0, 0)


    q := datastore.NewQuery("User")





    if _, err := q.GetAll(context, &users); err != nil {
        return "error"
    }



    leaders:=make([]Leader,0,len(users))
    for i:=0;i<len(users);i++{
        results := make([]Result, 0, 0)
        sum:=0
        getResults:=datastore.NewQuery("Result").Filter("Email =",users[i].Email)
        if _, err := getResults.GetAll(context, &results); err != nil {
           return "error"
        }
        for j:=0;j<len(results);j++ {
            sum+=results[j].Score
        }
        lead:=Leader{
            users[i].Name,
            sum,
        }
        leaders=append(leaders[:i],lead)


        
    }


    b, _ := json.Marshal(leaders)
    return string(b)

}


func FristLoad(context appengine.Context, replicas ...Load) string {
    c := make(chan string)
    loadReplica := func(i int) { c <- replicas[i](context) }
    for i := range replicas {
        go loadReplica(i)
    }
    return <-c
}



func appHandlerConc(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	
	status := r.FormValue("status")

	var result string
	var g1=playGame(status)
	var g2=playGame(status)
	var g3=playGame(status)

	c := make(chan string)
  go func() { c <- FirstGame(status, g1, g2,g3) } ()
  timeout := time.After(40 * time.Millisecond)

    select {
    case result = <-c:
        result = result
    case <-timeout:
        result = "timed out"
        break
    }

fmt.Fprint(w,result)



}

func appLeaderboardConc(w http.ResponseWriter, r *http.Request) {


    context := appengine.NewContext(r)
    w.Header().Set("Content-Type", "application/json")
    

    var result string
    var l1=doLoad(context)
    var l2=doLoad(context)
    var l3=doLoad(context)

    c := make(chan string)
  go func() { c <- FristLoad(context, l1, l2,l3) } ()
  timeout := time.After(500 * time.Millisecond)
  
    select {
    case result = <-c:
        result = result
    case <-timeout:
        result = "timed out"
        break
    }

fmt.Fprint(w,result)



}




func (g *Game) play()  (string , []string) {
  newStatus := ""

  sTableau := make([]string,0,9)

  for i, r := range g.Status {
    c := string(r)
    sTableau=append(sTableau[:i],c)
    newStatus+=c
}



if(g.Tableau[0][0]+g.Tableau[0][1]+g.Tableau[0][2]==2){
   for i:=0;i<3;i++{

      if g.Tableau[0][i]==0 {
         sTableau[i]="O"
         g.statusFromTableau(sTableau)
         g.makeTableau()
         return g.Status,sTableau
     }

 }
}

if(g.Tableau[1][0]+g.Tableau[1][1]+g.Tableau[1][2]==2){
   for i:=0;i<3;i++{

      if g.Tableau[1][i]==0 {
         sTableau[i+3]="O"
         g.statusFromTableau(sTableau)
         g.makeTableau()
         return g.Status,sTableau
     }

 }
}

if(g.Tableau[2][0]+g.Tableau[2][1]+g.Tableau[2][2]==2){
   for i:=0;i<3;i++{

      if g.Tableau[2][i]==0 {
         sTableau[i+6]="O"
         g.statusFromTableau(sTableau)
         g.makeTableau()
         return g.Status,sTableau
     }

 }
}


if(g.Tableau[0][0]+g.Tableau[1][1]+g.Tableau[2][2]==2){
   for i:=0;i<3;i++{

      if g.Tableau[i][i]==0 {
         sTableau[i+i]="O"
         g.statusFromTableau(sTableau)
         g.makeTableau()
         return g.Status,sTableau
     }

 }
}

if(g.Tableau[0][2]+g.Tableau[1][1]+g.Tableau[2][0]==2){
  j:=2
  p:=2
  for i:=0;i<3;i++ {

      if g.Tableau[i][j]==0 {
         sTableau[p*(i+1)]="O"
         g.statusFromTableau(sTableau)
         g.makeTableau()
         return g.Status,sTableau
     }
     j--

 }
}


if(g.Tableau[0][0]+g.Tableau[1][0]+g.Tableau[2][0]==2){
    j:=0
    for i:=0;i<3;i++{

        if g.Tableau[i][0]==0 {
            sTableau[i*j]="O"
            g.statusFromTableau(sTableau)
            g.makeTableau()
            return g.Status,sTableau
        }
        j+=3

    }
}


if(g.Tableau[0][1]+g.Tableau[1][1]+g.Tableau[2][1]==2){
    j:=1
    for i:=0;i<3;i++{

        if g.Tableau[i][1]==0 {
            sTableau[j]="O"
            g.statusFromTableau(sTableau)
            g.makeTableau()
            return g.Status,sTableau
        }
        j+=3

    }
}

if(g.Tableau[0][2]+g.Tableau[1][2]+g.Tableau[2][2]==2){
    j:=2
    for i:=0;i<3;i++{

        if g.Tableau[i][2]==0 {
            sTableau[j]="O"
            g.statusFromTableau(sTableau)
            g.makeTableau()
            return g.Status,sTableau
        }
        j+=3

    }
}

     //prima scelta
if sTableau[4]=="-" {

    sTableau[4]="O"
    g.statusFromTableau(sTableau)
    g.makeTableau()
    return g.Status,sTableau
} 


    //seconde scelte
if sTableau[0]=="-"{
   sTableau[0]="O"
   g.statusFromTableau(sTableau)
   g.makeTableau()
   return g.Status,sTableau
}

if sTableau[2]=="-" {
   sTableau[2]="O"
   g.statusFromTableau(sTableau)
   g.makeTableau()
   return g.Status,sTableau

} 
if sTableau[6]=="-" {
   sTableau[6]="O"
   g.statusFromTableau(sTableau)
   g.makeTableau()
   return g.Status,sTableau
}
if sTableau[8]=="-" {
  sTableau[8]="O"
  g.statusFromTableau(sTableau)
  g.makeTableau()
  return g.Status,sTableau
}

    //terze scelte
if sTableau[1]=="-"{
   sTableau[1]="O"
   g.statusFromTableau(sTableau)
   g.makeTableau()
   return g.Status,sTableau
} 
if sTableau[3]=="-" {
   sTableau[3]="O"
   g.statusFromTableau(sTableau)
   g.makeTableau()
   return g.Status,sTableau
} 
if sTableau[5]=="-" {
   sTableau[5]="O"
   g.statusFromTableau(sTableau)
   g.makeTableau()
   return g.Status,sTableau
}
if sTableau[7]=="-" {
  sTableau[7]="O"
  g.statusFromTableau(sTableau)
  g.makeTableau()
  return g.Status,sTableau
}


g.statusFromTableau(sTableau)
g.makeTableau()
return g.Status,sTableau


}


func (g *Game) statusFromTableau(sTableau []string) /*string*/ {

	g.Status=""
	for _, r := range sTableau {
		c := string(r)
       g.Status+=c
   }
    //return g.Status
}


func (g *Game) makeTableau() /*[][]int*/{


	
	buffArray := make([]string,0,9)
	x:=make([][]int,3) 
	
	for i := range x {
		x[i]=make([]int,3)
	}

	for i, r := range g.Status {
		c := string(r)
       buffArray=append(buffArray[:i],c)

   }

   firstRow:=buffArray[0:3]
   secondRow:=buffArray[3:6]
   thirdRow:=buffArray[6:9]


   for i,r := range firstRow{
      c:=string(r)
      switch(c){
         case "O" : x[0][i]=1
         break
         case "X" : x[0][i]=-1
         break
         case "-" : x[0][i]=0
         break

     }
 }

 for i,r := range secondRow{
  c:=string(r)
  switch(c){
     case "O" : x[1][i]=1
     break
     case "X" : x[1][i]=-1
     break
     case "-" : x[1][i]=0
     break

 }
}

for i,r := range thirdRow{
  c:=string(r)
  switch(c){
     case "O" : x[2][i]=1
     break
     case "X" : x[2][i]=-1
     break
     case "-" : x[2][i]=0
     break

 }
}

g.Tableau=x
    //return x

}




func commentKey (c appengine.Context) *datastore.Key{
	return datastore.NewKey(c,"User","default_user",0,nil)
}

func init() {
    http.HandleFunc("/", handler)

    http.HandleFunc("/app", appHandler)

    http.HandleFunc("/appConcurrent", appHandlerConc)


    http.HandleFunc("/appHistory", appHistory)

    http.HandleFunc("/appSave", appSave)

    http.HandleFunc("/appLeaderboard", appLeaderboard)

    http.HandleFunc("/appNewUser", appNewUser)

    http.HandleFunc("/appLeaderboardConc", appLeaderboardConc)



    


}


func appHistory(w http.ResponseWriter, r *http.Request) {


    var u User

    //fmt.Fprint(w, "hello i'm appHandler\n")
    context := appengine.NewContext(r)

    w.Header().Set("Content-Type", "application/json")
    
    status := r.FormValue("status")
    

    err := json.Unmarshal([]byte(status), &u)
    if err != nil {
        e,_:=json.Marshal(err)
        fmt.Fprint(w, string(e))
    }




    results := make([]Result, 0, 10)


    q := datastore.NewQuery("Result").Filter("Email =",u.Email).Limit(10)





    if _, err := q.GetAll(context, &results); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }






    b, _ := json.Marshal(results)
    fmt.Fprint(w, string(b))

}

func appSave(w http.ResponseWriter, r *http.Request) {


    var sc Result

    c := appengine.NewContext(r)

    w.Header().Set("Content-Type", "application/json")
    
    status := r.FormValue("status")

    err := json.Unmarshal([]byte(status), &sc)
    if err != nil {
        e,_:=json.Marshal(err)
        fmt.Fprint(w, string(e))
    }



    qU := datastore.NewQuery("User").Filter("Email =",sc.Email).Limit(1)



    
    for t := qU.Run(c); ; {
        var x User
        _, err := t.Next(&x)
        if err == datastore.Done {
            break
        }
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        if(x.Email==sc.Email){

            k := datastore.NewKey(c, "Result", time.Now().Format("2006-01-02T15:04:05.999999-07:00"), 0, nil)
            datastore.Put(c,k,&sc)
        }
    }

}


func appNewUser(w http.ResponseWriter, r *http.Request) {
    var u User

    c := appengine.NewContext(r)

    w.Header().Set("Content-Type", "application/json")
    
    status := r.FormValue("status")

    err := json.Unmarshal([]byte(status), &u)
    if err != nil {
        e,_:=json.Marshal(err)
        fmt.Fprint(w, string(e))
    }




    
    

    k := datastore.NewKey(c, "User", u.Email, 0, nil)
    datastore.Put(c,k,&u)
    


    sj,_:=json.Marshal(u)
    fmt.Fprint(w, string(sj))

    

}



func appLeaderboard(w http.ResponseWriter, r *http.Request) {



    context := appengine.NewContext(r)

    w.Header().Set("Content-Type", "application/json")

    fmt.Fprint(w, makeLeaderBoard(context))
    /*
    users := make([]User, 0, 0)


    q := datastore.NewQuery("User")





    if _, err := q.GetAll(context, &users); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }



    leaders:=make([]Leader,0,len(users))
    for i:=0;i<len(users);i++{
        results := make([]Result, 0, 0)
        sum:=0
        getResults:=datastore.NewQuery("Result").Filter("Email =",users[i].Email)
        if _, err := getResults.GetAll(context, &results); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        for j:=0;j<len(results);j++ {
            sum+=results[j].Score
        }
        lead:=Leader{
            users[i].Name,
            sum,
        }
        leaders=append(leaders[:i],lead)


        
    }


    b, _ := json.Marshal(leaders)
    fmt.Fprint(w, string(b))*/
    
}




func appHandler(w http.ResponseWriter, r *http.Request) {


	var g Game

	w.Header().Set("Content-Type", "application/json")
	
	status := r.FormValue("status")
	

	err := json.Unmarshal([]byte(status), &g)
	if err != nil {
		e,_:=json.Marshal(err)
		fmt.Fprint(w, string(e))
	}

	
	g.play()
	

	g.makeTableau()


	b, _ := json.Marshal(g)
	fmt.Fprint(w,string(b)+"\n")
	

}






func handler(w http.ResponseWriter, r *http.Request) {


    fmt.Fprint(w, "Not implemented")


}

