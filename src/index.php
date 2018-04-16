<?php

use ICal\ICal;

//Global absolute path
$appPath = realpath(dirname(__FILE__));
if (!preg_match('!/$!', $appPath)) {
    $appPath .= '/';
}

//Store in constants
define('APPLICATION_PATH', $appPath);
define('TIMEZONE', 'Europe/Berlin');

//Setup Timezone
$defaultTimeZone = new \DateTimeZone(TIMEZONE);
date_default_timezone_set(TIMEZONE);

//Include composer components
require $appPath . './vendor/autoload.php';

//Don't output errors
error_reporting(E_ALL);
ini_set('display_errors', 0);

//Make sure php is using utf as well as the output is recognized as utf8
mb_internal_encoding('UTF-8');


/**
 * Show a nice information overview page, if the parameters are not set
 * Also catch ppl trying to inject something over the parameters.
 */
if (!isset($_GET['pStud'], $_GET['pToken']) || !ctype_alnum($_GET['pStud']) || !ctype_alnum($_GET['pToken'])) {
    header('Content-Type: text/html; charset=UTF-8');
    include APPLICATION_PATH . 'about.php';
    die();
}

/**
 * Parse the file
 */
$calAddress = 'https://campus.tum.de/tumonlinej/ws/termin/ical?pStud=' . $_GET['pStud'] . '&pToken=' . $_GET['pToken'];
$iCal = new ICal($calAddress);
$allEvents = $iCal->events();

//Check if anything was received
if (empty($allEvents)) {
    die('Die parameter sind ung&uuml;ltig oder ein anderer Fehler ist aufgetreten');
}

//Remove dupes
CalProxy\handler::noDupes($allEvents);

//Create new object for outputting the new calender
$cal = new \Eluceo\iCal\Component\Calendar('TUM iCal Proxy');

// Create timezone rule object for Daylight Saving Time
$vTimezoneRuleDst = new \Eluceo\iCal\Component\TimezoneRule(\Eluceo\iCal\Component\TimezoneRule::TYPE_DAYLIGHT);
$vTimezoneRuleDst->setTzName('CEST');
$vTimezoneRuleDst->setDtStart(new \DateTime('1981-03-29 02:00:00', $defaultTimeZone));
$vTimezoneRuleDst->setTzOffsetFrom('+0100');
$vTimezoneRuleDst->setTzOffsetTo('+0200');
$dstRecurrenceRule = new \Eluceo\iCal\Property\Event\RecurrenceRule();
$dstRecurrenceRule->setFreq(\Eluceo\iCal\Property\Event\RecurrenceRule::FREQ_YEARLY);
$dstRecurrenceRule->setByMonth(3);
$dstRecurrenceRule->setByDay('-1SU');
$vTimezoneRuleDst->setRecurrenceRule($dstRecurrenceRule);
// Create timezone rule object for Standard Time
$vTimezoneRuleStd = new \Eluceo\iCal\Component\TimezoneRule(\Eluceo\iCal\Component\TimezoneRule::TYPE_STANDARD);
$vTimezoneRuleStd->setTzName('CET');
$vTimezoneRuleStd->setDtStart(new \DateTime('1996-10-27 03:00:00', $defaultTimeZone));
$vTimezoneRuleStd->setTzOffsetFrom('+0200');
$vTimezoneRuleStd->setTzOffsetTo('+0100');
$stdRecurrenceRule = new \Eluceo\iCal\Property\Event\RecurrenceRule();
$stdRecurrenceRule->setFreq(\Eluceo\iCal\Property\Event\RecurrenceRule::FREQ_YEARLY);
$stdRecurrenceRule->setByMonth(10);
$stdRecurrenceRule->setByDay('-1SU');
$vTimezoneRuleStd->setRecurrenceRule($stdRecurrenceRule);
// Create timezone definition and add rules
$vTimezone = new \Eluceo\iCal\Component\Timezone(TIMEZONE);
$vTimezone->addComponent($vTimezoneRuleDst);
$vTimezone->addComponent($vTimezoneRuleStd);
$cal->setTimezone($vTimezone);

//Event loop
foreach ($allEvents as $e) {
    $cal->addComponent(CalProxy\handler::cleanEvent($e));
}


header('Content-Type: text/calendar; charset=utf-8');
header('Content-Disposition: attachment; filename="cal.ics"');
echo $cal->render();
