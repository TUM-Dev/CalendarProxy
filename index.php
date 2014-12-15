<?php

require 'vendor/autoload.php';

error_reporting(E_ALL);
ini_set('display_errors', 1);

header('Content-Type: text/html; charset=UTF-8');
mb_internal_encoding('UTF-8');

function cleanEvent(&$e){
	$e['SUMMARY']=utf8_encode(str_replace(
	array('Tutorübungen', 'Grundlagen: ','Betriebssysteme und Systemsoftware', 'Einführung in die Informatik 2'), 
	array('TÜ','G','BS','INFO2'), 
	utf8_decode($e['SUMMARY'])));
	
	//Add missing fields if possible
	if(!isset($e['GEO'])){
		$e['GEO']='';
	}
	if(!isset($e['LOCATION'])){
		$e['LOCATION']='';
	}
	if(!isset($e['URL'])){
		$e['URL']='';
	}
	if(!isset($e['DESCRIPTION'])){
		$e['DESCRIPTION']='';
	}
	
}

function dumpMe($arr, $echo=true) {
    $str=str_replace(array("\n", ' '), array('<br/>', '&nbsp;'), print_r($arr, true)) . '<br/>';
    if($echo) {
        echo $str;
    }else{
        return $str;
    }
}

//Verify
if(!isset($_GET['pStud'],$_GET['pToken'])){
	$page=file_get_contents('about.html');
	die($page);
}

//Parse the file
$calAddr = 'https://campus.tum.de/tumonlinej/ws/termin/ical?pStud=' . $_GET['pStud'].'&pToken='.$_GET['pToken'];
$ical   = new ICal($calAddr);
$allEvents=$ical->events();

//Create new object
$cal = new \Eluceo\iCal\Component\Calendar('TUM iCal Proxy');

//output
foreach($allEvents as $e){
	$vEvent = new \Eluceo\iCal\Component\Event();
	
	//Process object
	cleanEvent($e);
	if(isset($_GET['debug'])){
		dumpMe($e);
	}
	
	//Create new and save it
	$vEvent
	->setUniqueId($e['UID'])
    ->setDtStart(new \DateTime($e['DTSTART']))
    ->setDtEnd(new \DateTime($e['DTEND']))
    ->setSummary($e['SUMMARY'])
	->setDescription($e['DESCRIPTION'])
	->setLocation($e['LOCATION'],$e['LOCATION'],$e['GEO'])
	->setUrl($e['URL']);
	
	$vEvent->setUseTimezone(true);
	$cal->addEvent($vEvent);
}


//Output if we are not debugging
if(!isset($_GET['debug'])){
	header('Content-Type: text/calendar; charset=utf-8');
	header('Content-Disposition: attachment; filename="cal.ics"');
	echo $cal->render();
}