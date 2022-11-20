<?php

use Eluceo\iCal\Domain\Entity\Calendar;
use Eluceo\iCal\Domain\Entity\TimeZone;
use ICal\ICal;
use DateTimeZone as PhpDateTimeZone;

//Global absolute path
$appPath = realpath(dirname(__FILE__));
if (!preg_match('!/$!', $appPath)) {
    $appPath .= '/';
}

//Store in constants
define('APPLICATION_PATH', $appPath);
define('TIMEZONE', 'Europe/Berlin');
ini_set('memory_limit', 8589934592);

//Setup Timezone
$defaultTimeZone = new \DateTimeZone(TIMEZONE);
date_default_timezone_set(TIMEZONE);

//Include composer components
require $appPath . '../vendor/autoload.php';
require $appPath . 'handler.php';

//Don't output errors
error_reporting(E_ALL);
ini_set('display_errors', 0);

//Make sure php is using utf as well as the output is recognized as utf8
mb_internal_encoding('UTF-8');


/**
 * Show a nice information overview page, if the parameters are not set
 * Also catch ppl trying to inject something over the parameters.
 */
if (!isset($_GET['pStud'], $_GET['pToken']) ) {
    header('Content-Type: text/html; charset=UTF-8');
    include APPLICATION_PATH . 'about.php';
    die();
}

// Catch people trying funny stuff
if(!ctype_alnum($_GET['pStud']) || !ctype_alnum($_GET['pToken'])){
    die('Ungültige ID angegeben!');
}

/**
 * Parse the file
 */
$calAddress = 'https://campus.tum.de/tumonlinej/ws/termin/ical?pStud=' . $_GET['pStud'] . '&pToken=' . $_GET['pToken'];
$iCal = new ICal($calAddress);
$allEvents = $iCal->events();

//Check if anything was received
if (empty($allEvents)) {
    die('Die parameter sind ungültig oder ein anderer Fehler ist aufgetreten. Es konnten keine Termine gefunden werden.');
}

//Remove dupes
CalProxy\handler::noDupes($allEvents);

// Do the transformation
$handler =  new CalProxy\handler();
$newEvents = [];
foreach ($allEvents as $e) {
    $newEvents[] = $handler->cleanEvent($e);
}

//Create new object for outputting the new calendar
$calendar = new Calendar($newEvents);

// 3. Transform domain entity into an iCalendar component
$componentFactory = new Eluceo\iCal\Presentation\Factory\CalendarFactory();
$calendarComponent = $componentFactory->createCalendar($calendar);

header('Content-Type: text/calendar; charset=utf-8');
header('Content-Disposition: attachment; filename="cal.ics"');
echo $calendarComponent;
