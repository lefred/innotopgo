# Innotop Go

Innotop for MySQL 8 written in Go

Project started to learn Go and doing something useful (I hope).

Additionally the official Innotop written in Perl became very hard to maintain.

# Screenshots

<table class="table">
  <thead>
    <tr>
      <th scope="col" width="1000px">Main Processlist Screen</th>
      <th scope="col" width="1000px">InnoDB Dashboard</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>
        <img src="https://user-images.githubusercontent.com/609675/113749711-3afc1c00-970a-11eb-8ace-ccd0e38cd443.png" width="100%" alt="Main Processlist Screen">
      </td>
      <td>
        <img src="https://user-images.githubusercontent.com/609675/114268187-249eda80-9a00-11eb-80ff-5aaebf378d78.png" width="100%" alt="InnoDB Dashboard">
      </td>
    </tr>
  </tbody>
</table>

<table class="table">
  <thead>
    <tr>
      <th scope="col" width="1000px">Memory Dashboard</th>
      <th scope="col" width="1000px">Error Log Dashboard</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>
        <img src="https://user-images.githubusercontent.com/609675/114268174-1486fb00-9a00-11eb-9264-55486d69d582.png" width="100%" alt="Memory Dashboard">
      </td>
      <td>
        <img src="https://user-images.githubusercontent.com/609675/125663301-9541a71b-1fd2-45d4-8469-ff3c957e61ae.png" width="100%" alt="Error Log Dashboard">
      </td>
    </tr>
  </tbody>
</table>

## Locking Info

<table class="table">
  <tbody>
    <tr>
      <td>
        <img src="https://user-images.githubusercontent.com/609675/125854001-6f8f33b9-7095-42b9-89de-593a4b119f41.png" width="100%" alt="Locking Info">
      </td>
      <td>
        <img src="https://user-images.githubusercontent.com/609675/125854013-bb358762-3db1-4b3a-9c5f-f26ae2a070e8.png" width="100%" alt="Locking Info">
      </td>
    </tr>
  </tbody>
</table>

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
