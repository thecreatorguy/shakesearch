# Tim Johnson's ShakeSearch
Hi, I'm Tim Johnson, a college student at RIT with 2 semesters remaining,
Spring 2021 and Fall 2021. With only a year left until my education is
complete, I think it's just about early enough to start looking into long
term careers. Not only does this job appeal to me because the company looks
like it's a great place to work, and the main technologies at play are
exactly those that I want to be involved with, but this "screening" 
process of a project allows me to flex/put into practice my current 
skills (Go) and also learn more about the technologies I want to become
more involved with (React). Pulley, if this ever sees your eyes, thank
you for this opportunity!

## Plan
This readme will be a full document of my thought processes and planning-
none if it will be edited, only new entries will be added, so as to see
the process clearly as it evolves. While I do have confidence in my skills,
I believe it's best to keep an open mind and to keep looking for more things
to improve upon.

### Entry 1- 1:35AM December 17th, 2020
The first things on my mind for improvements are the things mentioned in
previous readme: the case sensitivity of the search (which proved immediately
annoying when I tried to search "hamlet" instead of "Hamlet"), and the need
for a complete match for searching. Later, though- it's getting late!

### Entry 2- 2:49PM December 17th, 2020
I've now done some research on suffix arrays, because previously I did not
know how they worked. They are very cool, especially because their lookup
efficiency is O(n) and it is in-memory! Unfortunately, it looks like they
might not be up for the job of full text search because of its need for exact
matches. It isn't to the word level, though- for example, "hamle", "amlet",
and "hamlet" all return similar results- so it is probably good enough for now.
The biggest problem I see after case insensitivity is that the entire array
is displayed for the search, making it impossible to navigate. For this,
the results must be paginated, so that will be my first objective.

### Entry 3- 5:25PM December 17th, 2020
Pagination is now handled. As I was working on it, several new thoughts came
to mind:

1. There isn't enough information. Typically, when searching through
Shakespeare, one would like to know certain metadata like what play it is
from, what act, scene, and line numbers it is located at. This is a parsing
problem, because all of that data is available in the full text, but simply
reading the file doesn't give that to us.

2. It's not pretty, obviously. My current thought is to make it look similar
to the react prompt, but have the left side be the list of results, and the
right side be a full preview of the selected result.

3. The lack of "approximate search" is a real problem- when searching through
Shakespeare, most people won't have already memorized Shakespeare, and most
will probably be searching for more than just a word. How to do this while
keeping a fast search will probably be the biggest challenge.

After completing these three features, I think the project will probably
be complete.

### Entry 4- 12:28AM December 20th, 2020
It's clear that the source data is unedited. There are over 30 different
works, so I'm not going to create 30 different text files for each work,
although that would be somewhat feasible to do. Instead, I'm going to
use a unicode character not found in the text and place it before each
title, and delete all meta text like copyright and table of contents-
when searching for Shakespeare, people aren't looking for this. The
unicode character will be "🙂", and will act as a bookmark.