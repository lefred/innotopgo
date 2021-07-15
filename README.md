# Innotop Go
Innotop for MySQL 8 written in Go

Project started to learn Go and doing something useful (I hope).

Additionally the official Innotop written in Perl became very hard to maintain.

## Main Processlist Screen

![Screenshot from 2021-04-06 18-57-56](https://user-images.githubusercontent.com/609675/113749711-3afc1c00-970a-11eb-8ace-ccd0e38cd443.png)

## InnoDB Dashboard

![Screenshot from 2021-04-07 23-20-20](https://user-images.githubusercontent.com/609675/114268187-249eda80-9a00-11eb-80ff-5aaebf378d78.png)


## Memory Dashboard

![Screenshot from 2021-04-10 13-16-26](https://user-images.githubusercontent.com/609675/114268174-1486fb00-9a00-11eb-9264-55486d69d582.png)

## Error Log Dashboard - in dev

![Selection_084](https://user-images.githubusercontent.com/609675/125663301-9541a71b-1fd2-45d4-8469-ff3c957e61ae.png)

## Locking Info - in dev

![Screenshot from 2021-07-15 22-20-45](https://user-images.githubusercontent.com/609675/125854001-6f8f33b9-7095-42b9-89de-593a4b119f41.png)

![Screenshot from 2021-07-15 22-21-00](https://user-images.githubusercontent.com/609675/125854013-bb358762-3db1-4b3a-9c5f-f26ae2a070e8.png)


## Demo

Demo (0.1.1) on MacOS (thank you @datacharmer):

![innotopgo](https://user-images.githubusercontent.com/609675/113839514-08950200-9790-11eb-8cc6-449250909acb.gif)


## Connect

```bash
    ./innotopgo mysql://<username>:<password>@<host>:3306
```

example:

```bash
    ./innotopgo mysql://root:password@localhost:3306
```

## Help

Press <kbd>?</kbd> within *innotopgo* application.
