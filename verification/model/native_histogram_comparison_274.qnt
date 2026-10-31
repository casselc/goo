// Public comparison display/effects, not private source or collection authority.
module native_histogram_comparison_274 {
 const MAX_POSTS:int
 var posts:int
 var phase:int
 action init=all{posts'=0,phase'=0}
 action compare=all{phase==0,posts'=posts,phase'=1}
 action count=all{phase==1,posts'=posts,phase'=2}
 action share=all{phase==2,posts'=posts,phase'=3}
 action saved=all{phase==3,posts'=posts,phase'=4}
 action withheld=all{phase==4,posts'=posts,phase'=5}
 action complete=all{phase==5,posts'=posts,phase'=6}
 val safety=posts<=MAX_POSTS and posts>=0 and phase>=0 and phase<=6
}
