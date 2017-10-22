<?php

use ICal\ICal;

//Global absolute path
$appPath = realpath(dirname(__FILE__));
if (!preg_match('!/$!', $appPath)) {
    $appPath .= '/';
}

//Store in constants
define('APPLICATION_PATH', $appPath);
define('PATH_HEADER', APPLICATION_PATH . '../header.html');
define('PATH_FOOTER', APPLICATION_PATH . '../footer.html');
define('PATH_ABOUT', APPLICATION_PATH . 'about.html');
define('TIMEZONE', 'Europe/Berlin');

//Setup Timezone
$defaultTimeZone = new \DateTimeZone(TIMEZONE);
date_default_timezone_set(TIMEZONE);

//Include composer components
require $appPath . 'vendor/autoload.php';

//Secruity thingy: Comment this out to enable debugging
unset($_GET['debug']);

//Only output errors if debugging
if (isset($_GET['debug'])) {
    error_reporting(E_ALL);
    ini_set('display_errors', 1);
} else {
    error_reporting(0);
    ini_set('display_errors', 0);
}

//Make sure php is using utf as well as the output is recognized as utf8
header('Content-Type: text/html; charset=UTF-8');
mb_internal_encoding('UTF-8');


/**
 * Show a nice information overview page, if the parameters are not set
 * Also catch ppl trying to inject something over the parameters.
 */
if (!isset($_GET['pStud'], $_GET['pToken']) || !ctype_alnum($_GET['pStud']) || !ctype_alnum($_GET['pToken'])) {
    if (file_exists(PATH_HEADER) && file_exists(PATH_FOOTER)) {
        $page = file_get_contents(PATH_HEADER) . str_replace('%HOST%', $_SERVER['SERVER_NAME'] . '/' . basename(__DIR__), file_get_contents(PATH_ABOUT)) . file_get_contents(PATH_FOOTER);
    } else {
        $page = str_replace('%HOST%', $_SERVER['SERVER_NAME'], file_get_contents(PATH_ABOUT));
    }
    die($page);
}

/**
 * Parse the file
 */
$calAddress = 'https://campus.tum.de/tumonlinej/ws/termin/ical?pStud=' . $_GET['pStud'] . '&pToken=' . $_GET['pToken'];
$iCal = new ICal($calAddress);
$allEvents = $iCal->events();

//Check if anything was received
if (empty($allEvents)) {
    die('Ihre parameter sind ung&uuml;ltig oder ein anderer Fehler ist aufgetreten');
}

//Remove dupes
CalProxy\handler::noDupes($allEvents);

//Create new object for outputting the new calender
$cal = new \Eluceo\iCal\Component\Calendar('TUM iCal Proxy');
$cal->setTimezone(new \Eluceo\iCal\Component\Timezone(TIMEZONE));

//Event loop
foreach ($allEvents as $e) {
    $cal->addComponent(CalProxy\handler::cleanEvent($e));
}


//Output if we are not debugging
if (!isset($_GET['debug'])) {
    header('Content-Type: text/calendar; charset=utf-8');
    header('Content-Disposition: attachment; filename="cal.ics"');
    echo $cal->render();
}
