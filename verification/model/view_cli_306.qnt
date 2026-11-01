module view_cli_306 {
 const MAX_READS:int
 var phase:int
 var validations:int
 var refusals:int
 var imports:int
 var reads:int
 var saves:int
 var expands:int
 action init=all{phase'=0,validations'=0,refusals'=0,imports'=0,reads'=0,saves'=0,expands'=0}
 action validateChecked=all{phase==0,phase'=1,validations'=1,refusals'=0,imports'=0,reads'=0,saves'=0,expands'=0}
 action refuseInvalid=all{phase==1,phase'=2,validations'=1,refusals'=1,imports'=0,reads'=0,saves'=0,expands'=0}
 action validateScoped=all{phase==2,phase'=3,validations'=2,refusals'=1,imports'=0,reads'=0,saves'=0,expands'=0}
 action cliImport=all{phase==3,phase'=4,validations'=2,refusals'=1,imports'=1,reads'=0,saves'=0,expands'=0}
 action catalog=all{phase==4,phase'=5,validations'=2,refusals'=1,imports'=1,reads'=0,saves'=0,expands'=0}
 action open=all{phase==5,phase'=6,validations'=2,refusals'=1,imports'=1,reads'=0,saves'=0,expands'=0}
 action read=all{phase==6,phase'=7,validations'=2,refusals'=1,imports'=1,reads'=1,saves'=0,expands'=0}
 action destination=all{phase==7,phase'=8,validations'=2,refusals'=1,imports'=1,reads'=5,saves'=0,expands'=0}
 action apply=all{phase==8,phase'=9,validations'=2,refusals'=1,imports'=1,reads'=5,saves'=0,expands'=1}
 action save=all{phase==9,phase'=10,validations'=2,refusals'=1,imports'=1,reads'=5,saves'=1,expands'=1}
 val safety=validations>=0 and validations<=2 and refusals>=0 and refusals<=1 and imports>=0 and imports<=1 and reads>=0 and reads<=MAX_READS and saves>=0 and saves<=1 and expands>=0 and expands<=1 and phase>=0 and phase<=10
}
